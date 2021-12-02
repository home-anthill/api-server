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
	"api-server/api"
	"api-server/api/oauth"
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

var auth *api.Auth
var homes *api.Homes
var devices *api.Devices
var profiles *api.Profiles
var register *api.Register
var collectionProfiles *mongo.Collection

func main() {
	logger := InitLogger()
	defer logger.Sync()

	logger.Info("Starting application...)")

	//Load the .env file
	err := godotenv.Load(".env")
	if err != nil {
		logger.Error("failed to load the env file")
	}

	// read ENV property from .env
	if os.Getenv("ENV") == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	logger.Info("ENVIRONMENT = " + os.Getenv("ENV"))
	logger.Info("HTTP PORT = " + os.Getenv("HTTP_PORT"))
	logger.Info("GRPC PORT = " + os.Getenv("GRPC_PORT"))

	ctx := context.Background()

	logger.Info("Connecting to MongoDB...")
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017/"))
	if err = client.Ping(context.TODO(), readpref.Primary()); err != nil {
		log.Fatal(err)
	}
	logger.Info("Connected to MongoDB")

	collectionProfiles = client.Database(DbName).Collection("profiles")
	collectionHomes := client.Database(DbName).Collection("homes")
	collectionDevices := client.Database(DbName).Collection("devices")

	auth = api.NewAuth(ctx, logger, collectionProfiles)
	homes = api.NewHomes(ctx, logger, collectionHomes, collectionProfiles)
	devices = api.NewDevices(ctx, logger, collectionDevices, collectionProfiles, collectionHomes)
	profiles = api.NewProfiles(ctx, logger, collectionProfiles)
	register = api.NewRegister(ctx, logger, collectionDevices, collectionProfiles)

	amqpSubscriber.InitAmqpSubscriber()

	hubInstance := ws.GetInstance()
	go hubInstance.Run()

	router := gin.Default()

	// implement websocket to receive realtime events from rabbitmq via amqp
	// this service should be protected by auth
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
	oauth.Setup(redirectURL, credFile, scopes, secret, collectionProfiles)
	sessionName := "session"
	router.Use(oauth.Session(sessionName))

	router.GET("/api/login", oauth.GetLoginURL)

	router.POST("/api/register", register.PostRegister)

	// protected url group
	authorized := router.Group("/auth")
	authorized.Use(oauth.OauthAuth())
	{
		authorized.GET("", auth.LoginCallback)
	}

	// protected url group
	private := router.Group("/api")
	private.Use(auth.JWTMiddleware())
	{
		private.GET("/homes", homes.GetHomes)
		private.POST("/homes", homes.PostHome)
		private.PUT("/homes/:id", homes.PutHome)
		private.DELETE("/homes/:id", homes.DeleteHome)
		private.GET("/homes/:id/rooms", homes.GetRooms)
		private.POST("/homes/:id/rooms", homes.PostRoom)
		private.PUT("/homes/:id/rooms/:rid", homes.PutRoom)
		private.DELETE("/homes/:id/rooms/:rid", homes.DeleteRoom)

		private.GET("/profile", profiles.GetProfile)
		private.POST("/profiles/:id/tokens", profiles.PostProfilesToken)

		private.GET("/devices", devices.GetDevices)
		private.DELETE("/devices/:id", devices.DeleteDevice)

		private.GET("/devices/:id/values", devices.GetValuesDevice)
		private.POST("/devices/:id/values/onoff", devices.PostOnOffDevice)
		private.POST("/devices/:id/values/temperature", devices.PostTemperatureDevice)
		private.POST("/devices/:id/values/mode", devices.PostModeDevice)
		private.POST("/devices/:id/values/fanmode", devices.PostFanModeDevice)
		private.POST("/devices/:id/values/fanspeed", devices.PostFanSpeedDevice)
	}

	port := os.Getenv("HTTP_PORT")

	err = router.Run(":" + port)
	if err != nil {
		logger.Error("Cannot start HTTP server", err)
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
