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
	"air-conditioner/github"
	"air-conditioner/handlers"
	"context"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"log"
)

var authHandler *handlers.AuthHandler
var homesHandler *handlers.HomesHandler
var acsHandler *handlers.ACsHandler
var profilesHandler *handlers.ProfilesHandler

func init() {
	ctx := context.Background()
	log.Println("Connecting to MongoDB...")
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017/"))
	if err = client.Ping(context.TODO(), readpref.Primary()); err != nil {
		log.Fatal(err)
	}
	log.Println("Connected to MongoDB")
	collectionUsers := client.Database("airConditionerDb").Collection("users")
	collectionHomes := client.Database("airConditionerDb").Collection("homes")
	collectionACs := client.Database("airConditionerDb").Collection("airconditioners")

	authHandler = handlers.NewAuthHandler(ctx, collectionUsers)
	homesHandler = handlers.NewHomesHandler(ctx, collectionHomes)
	acsHandler = handlers.NewACsHandler(ctx, collectionACs)
	profilesHandler = handlers.NewProfilesHandler(ctx, collectionUsers)
}

func main() {
	router := gin.Default()

	// - No origin allowed by default
	// - GET,POST, PUT, HEAD methods
	// - Credentials share disabled
	// - Preflight requests cached for 12 hours
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:3000", "http://localhost:8080"}
	// config.AllowOrigins == []string{"http://google.com", "http://facebook.com"}
	router.Use(cors.New(config))

	router.Use(static.Serve("/", static.LocalFile("./public", true)))
	router.Use(static.Serve("/postlogin", static.LocalFile("./public", true)))

	// init settings for github auth
	redirectURL := "http://localhost:8080/auth/"
	credFile := "./credentials.json"
	// You have to select your own scope from here -> https://developer.github.com/v3/oauth/#scopes
	scopes := []string{"repo"}
	secret := []byte("secret")
	github.Setup(redirectURL, credFile, scopes, secret)
	sessionName := "goquestsession"
	router.Use(github.Session(sessionName))

	router.GET("/login", github.GetLoginURLHandler)

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

		private.POST("/profiles/:id/tokens", profilesHandler.PostProfilesTokenHandler)

		private.GET("/airconditioners", acsHandler.GetACsHandler)
		private.POST("/airconditioners", acsHandler.PostACHandler)
		private.PUT("/airconditioners/:id", acsHandler.PutACHandler)
		private.DELETE("/airconditioners/:id", acsHandler.DeleteACHandler)
	}

	err := router.Run(":8080")
	if err != nil {
		panic(err)
	}
}
