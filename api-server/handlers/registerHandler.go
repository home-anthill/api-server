package handlers

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"net/http"
	"os"
	"time"

	"api-server/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/net/context"

	pb "api-server/register"
	"google.golang.org/grpc"
)

type DeviceRequest struct {
	//swagger:ignore
	Mac          string `json:"mac"`
	Name         string `json:"name"`
	Manufacturer string `json:"manufacturer"`
	Model        string `json:"model"`
	Type         string `json:"type"`
	APIToken     string `json:"apiToken"`
}

type RegisterHandler struct {
	collection         *mongo.Collection
	collectionProfiles *mongo.Collection
	ctx                context.Context
	grpcTarget         string
}

func NewRegisterHandler(ctx context.Context, collection *mongo.Collection, collectionProfiles *mongo.Collection) *RegisterHandler {
	grpcPort := os.Getenv("GRPC_PORT")

	return &RegisterHandler{
		collection:         collection,
		collectionProfiles: collectionProfiles,
		ctx:                ctx,
		grpcTarget:         "localhost:" + grpcPort,
	}
}

func (handler *RegisterHandler) PostRegisterHandler(c *gin.Context) {
	// receive a payload from devices with
	var registerBody DeviceRequest
	if err := c.ShouldBindJSON(&registerBody); err != nil {
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	// search if profile token exists and retrieve profile
	var profileFound models.Profile
	errProfile := handler.collectionProfiles.FindOne(handler.ctx, bson.M{
		"apiToken": registerBody.APIToken,
	}).Decode(&profileFound)
	if errProfile != nil {
		fmt.Println("Cannot find profile with that apiToken", errProfile)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot register, profile token missing or not valid"})
		return
	}

	// search and skip db add if already exists
	var device models.Device
	err := handler.collection.FindOne(handler.ctx, bson.M{
		"mac": registerBody.Mac,
	}).Decode(&device)
	if err == nil {
		// if err == nil => ac found in db (already exists)
		// skip register process returning "already registered"
		c.JSON(http.StatusOK, gin.H{"message": "already registered"})
		return
	}

	device = models.Device{}
	device.ID = primitive.NewObjectID()
	device.Mac = registerBody.Mac
	device.Name = registerBody.Name
	device.Manufacturer = registerBody.Manufacturer
	device.Model = registerBody.Model
	device.CreatedAt = time.Now()
	device.ModifiedAt = time.Now()

	_, err2 := handler.collection.InsertOne(handler.ctx, device)
	if err2 != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while inserting a new ac"})
		return
	}

	// push AC.ID to profile.devices
	result, errUpd := handler.collectionProfiles.UpdateOne(
		handler.ctx,
		bson.M{"_id": profileFound.ID},
		bson.M{"$push": bson.M{"devices": device.ID}},
	)

	fmt.Println("result: ", result)
	fmt.Println("errUpd: ", errUpd)

	// Set up a connection to the server.
	conn, err := grpc.Dial(handler.grpcTarget, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	client := pb.NewRegistrationClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := client.Register(ctx, &pb.RegisterRequest{
		Id:             device.ID.Hex(),
		Mac:            device.Mac,
		Name:           device.Name,
		Manufacturer:   device.Manufacturer,
		Model:          device.Model,
		ProfileOwnerId: profileFound.ID.Hex(),
		ApiToken:       profileFound.ApiToken,
	})
	if err != nil {
		log.Fatalf("could not register: %v", err)
	}
	fmt.Println("Register status: ", r.GetStatus())
	fmt.Println("Register message: ", r.GetMessage())

	c.JSON(http.StatusOK, device)
}
