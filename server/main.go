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
	handlers "air-conditioner/handlers"
	"context"
	"github.com/gin-gonic/gin"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"log"
)

var homesHandler *handlers.HomesHandler
var acsHandler *handlers.ACsHandler

func init() {
	ctx := context.Background()
	log.Println("Connecting to MongoDB...")
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017/"))
	if err = client.Ping(context.TODO(), readpref.Primary()); err != nil {
		log.Fatal(err)
	}
	log.Println("Connected to MongoDB")
	collectionHomes := client.Database("airConditionerDb").Collection("homes")
	collectionACs := client.Database("airConditionerDb").Collection("airconditioners")

	homesHandler = handlers.NewHomesHandler(ctx, collectionHomes)
	acsHandler = handlers.NewACsHandler(ctx, collectionACs)
}

func main() {
	router := gin.Default()

	router.GET("/homes", homesHandler.GetHomesHandler)
	router.POST("/homes", homesHandler.PostHomeHandler)
	router.PUT("/homes/:id", homesHandler.PutHomeHandler)
	router.DELETE("/homes/:id", homesHandler.DeleteHomeHandler)
	router.GET("/homes/:id/rooms/", homesHandler.GetRoomsHandler)
	router.POST("/homes/:id/rooms/", homesHandler.PostRoomHandler)
	router.PUT("/homes/:id/rooms/:rid", homesHandler.PutRoomHandler)
	router.DELETE("/homes/:id/rooms/:rid", homesHandler.DeleteRoomHandler)

	router.GET("/airconditioners", acsHandler.GetACsHandler)
	router.POST("/airconditioners", acsHandler.PostACHandler)
	router.PUT("/airconditioners/:id", acsHandler.PutACHandler)
	router.DELETE("/airconditioners/:id", acsHandler.DeleteACHandler)

	router.Run(":8080")
}
