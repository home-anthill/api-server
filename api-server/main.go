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
	amqpSubscriber "air-conditioner/amqp-subscriber"
	"air-conditioner/github"
	"air-conditioner/handlers"
	"air-conditioner/ws"
	"context"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"log"
	"path"
	"path/filepath"
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
	amqpSubscriber.InitAmqpSubscriber()

	hubInstance := ws.GetInstance()
	go hubInstance.Run()

	router := gin.Default()

	// implement websocket to receive realtime events from rabbitmq via amqp
	// this service should be protected by authHandler
	router.GET("/ws", func (c *gin.Context) {
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
	github.Setup(redirectURL, credFile, scopes, secret)
	sessionName := "goquestsession"
	router.Use(github.Session(sessionName))

	router.GET("/api/login", github.GetLoginURLHandler)

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

		private.GET("/airconditioners", acsHandler.GetACsHandler)
		private.POST("/airconditioners", acsHandler.PostACHandler)
		private.PUT("/airconditioners/:id", acsHandler.PutACHandler)
		private.DELETE("/airconditioners/:id", acsHandler.DeleteACHandler)
	}

	err := router.Run(":8082")
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