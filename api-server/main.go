// Air Conditioner API
//
// Air Conditioner control system APIs.
//
//	Schemes: http
//  Host: localhost:3000
//	BasePath: /
//	Version: 1.0.0
//	Contact: Stefano Cappa <stefano.cappa.ks89@gmail.com> https://github.com/Ks89
//
//	Consumes:
//	- application/json
//
//	Produces:
//	- application/json
// swagger:meta
package main

import (
	amqpSubscriber "api-server/amqp-subscriber"
	"api-server/github"
	"api-server/handlers"
	"api-server/ws"
	"context"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"log"
	"os"
	"path"
	"path/filepath"
)

const DbName = "api-server"

var authHandler *handlers.AuthHandler
var homesHandler *handlers.HomesHandler
var devicesHandler *handlers.DevicesHandler
var profilesHandler *handlers.ProfilesHandler
var registerHandler *handlers.RegisterHandler
var collectionProfiles *mongo.Collection

func init() {
	//Load the .env file
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("error: failed to load the env file")
	}

	// read ENV property from .env
	if os.Getenv("ENV") == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	ctx := context.Background()
	log.Println("Connecting to MongoDB...")
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017/"))
	if err = client.Ping(context.TODO(), readpref.Primary()); err != nil {
		log.Fatal(err)
	}
	log.Println("Connected to MongoDB")
	collectionProfiles = client.Database(DbName).Collection("profiles")
	collectionHomes := client.Database(DbName).Collection("homes")
	collectionDevices := client.Database(DbName).Collection("devices")

	authHandler = handlers.NewAuthHandler(ctx, collectionProfiles)
	homesHandler = handlers.NewHomesHandler(ctx, collectionHomes, collectionProfiles)
	devicesHandler = handlers.NewDevicesHandler(ctx, collectionDevices, collectionProfiles, collectionHomes)
	profilesHandler = handlers.NewProfilesHandler(ctx, collectionProfiles)
	registerHandler = handlers.NewRegisterHandler(ctx, collectionDevices, collectionProfiles)
}

func main() {
	amqpSubscriber.InitAmqpSubscriber()

	hubInstance := ws.GetInstance()
	go hubInstance.Run()

	router := gin.Default()

	// implement websocket to receive realtime events from rabbitmq via amqp
	// this service should be protected by authHandler
	router.GET("/ws", func(c *gin.Context) {
		ws.ServeWs(c.Writer, c.Request)
	})

	// - No origin allowed by default
	// - GET,POST, PUT, HEAD methods
	// - Credentials share disabled
	// - Preflight requests cached for 12 hours
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:3000", "http://localhost:8082"}
	// config.AllowOrigins == []string{"http://google.com", "http://facebook.com"}
	router.Use(cors.New(config))

	router.Use(gzip.Gzip(gzip.DefaultCompression))

	// GIN is terrible with SPA, because you can configure static.serve
	// but if you refresh the SPA it will return an error and you cannot add something like /*
	// The only way is to manage this manually passing the filename in case it's a file, otherwise it must redirect
	// to the index.html page
	//router.Use(static.Serve("/", static.LocalFile("./public", false)))
	router.NoRoute(func(c *gin.Context) {
		dir, file := path.Split(c.Request.RequestURI)
		ext := filepath.Ext(file)
		allowedExts := []string{".html", ".htm", ".js", ".css", ".json", ".txt", ".jpeg", ".jpg", ".png", ".ico", ".map", ".svg"}
		_, found := Find(allowedExts, ext)
		if found {
			c.File("./public" + path.Join(dir, file))
		} else {
			c.File("./public/index.html")
		}
	})

	// init settings for github auth
	redirectURL := "http://localhost:8082/auth/"
	credFile := "./credentials.json"
	// You have to select your own scope from here -> https://developer.github.com/v3/oauth/#scopes
	scopes := []string{"repo"}
	secret := []byte("secret")
	github.Setup(redirectURL, credFile, scopes, secret, collectionProfiles)
	sessionName := "session"
	router.Use(github.Session(sessionName))

	router.GET("/api/login", github.GetLoginURLHandler)

	router.POST("/api/register", registerHandler.PostRegisterHandler)

	// protected url group
	authorized := router.Group("/auth")
	authorized.Use(github.Auth())
	{
		authorized.GET("", authHandler.LoginCallbackHandler)
	}

	// protected url group
	private := router.Group("/api")
	private.Use(authHandler.JWTMiddleware())
	{
		private.GET("/homes", homesHandler.GetHomesHandler)
		private.POST("/homes", homesHandler.PostHomeHandler)
		private.PUT("/homes/:id", homesHandler.PutHomeHandler)
		private.DELETE("/homes/:id", homesHandler.DeleteHomeHandler)
		private.GET("/homes/:id/rooms", homesHandler.GetRoomsHandler)
		private.POST("/homes/:id/rooms", homesHandler.PostRoomHandler)
		private.PUT("/homes/:id/rooms/:rid", homesHandler.PutRoomHandler)
		private.DELETE("/homes/:id/rooms/:rid", homesHandler.DeleteRoomHandler)

		private.GET("/profile", profilesHandler.GetProfileHandler)
		private.POST("/profiles/:id/tokens", profilesHandler.PostProfilesTokenHandler)

		private.GET("/devices", devicesHandler.GetDevicesHandler)
		private.DELETE("/devices/:id", devicesHandler.DeleteDeviceHandler)

		private.GET("/devices/:id/values", devicesHandler.GetValuesDeviceHandler)
		private.POST("/devices/:id/values/onoff", devicesHandler.PostOnOffDeviceHandler)
		private.POST("/devices/:id/values/temperature", devicesHandler.PostTemperatureDeviceHandler)
		private.POST("/devices/:id/values/mode", devicesHandler.PostModeDeviceHandler)
		private.POST("/devices/:id/values/fanmode", devicesHandler.PostFanModeDeviceHandler)
		private.POST("/devices/:id/values/fanspeed", devicesHandler.PostFanSpeedDeviceHandler)
	}

	port := os.Getenv("HTTP_PORT")

	err := router.Run(":" + port)
	if err != nil {
		panic(err)
	}
}

func Find(slice []string, val string) (int, bool) {
	for i, item := range slice {
		if item == val {
			return i, true
		}
	}
	return -1, false
}
