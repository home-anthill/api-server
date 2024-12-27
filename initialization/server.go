package initialization

import (
	"api-server/api"
	"api-server/utils"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	limits "github.com/gin-contrib/size"
	"github.com/gin-gonic/contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	"os"
	"path"
	"path/filepath"
)

var oauthGithub *api.GitHub
var auth *api.Auth
var homes *api.Homes
var devices *api.Devices
var assignDevices *api.AssignDevice
var devicesValues *api.DevicesValues
var profiles *api.Profiles
var register *api.Register
var fcmToken *api.FCMToken
var online *api.Online
var keepAlive *api.KeepAlive

var oauthCallbackURL string
var oauthScopes = []string{"repo"} //https://developer.github.com/v3/oauth/#scopes

// SetupRouter function
func SetupRouter(logger *zap.SugaredLogger) (*gin.Engine, cookie.Store) {
	port := os.Getenv("HTTP_PORT")
	httpServer := os.Getenv("HTTP_SERVER")
	oauthCallback := os.Getenv("OAUTH_CALLBACK")

	// 1. init oauthCallbackURL based on httpOrigin
	oauthCallbackURL = oauthCallback
	logger.Info("SetupRouter - oauthCallbackURL is = " + oauthCallbackURL)

	httpOrigin := httpServer + ":" + port
	logger.Info("SetupRouter - httpOrigin is = " + httpOrigin)

	// 2. init GIN
	router := gin.Default()
	// 3. init session
	cookieStore := cookie.NewStore([]byte("secret"))
	router.Use(sessions.Sessions("session", cookieStore))
	// 4. apply compression
	router.Use(gzip.Gzip(gzip.DefaultCompression))

	// 5. fix a max POST payload size
	logger.Info("SetupRouter - set mac POST payload size")
	router.Use(limits.RequestSizeLimiter(1024 * 1024))

	// 6. Configure CORS
	// - No origin allowed by default
	// - GET,POST, PUT, HEAD methods
	// - Credentials share disabled
	// - Preflight requests cached for 12 hours
	if os.Getenv("HTTP_CORS") == "true" {
		logger.Warn("SetupRouter - CORS enabled and httpOrigin is = " + httpOrigin)
		config := cors.DefaultConfig()
		config.AllowOrigins = []string{
			"http://" + os.Getenv("INTERNAL_CLUSTER_PATH"),
			"http://" + os.Getenv("INTERNAL_CLUSTER_PATH") + ":80",
			"https://" + os.Getenv("INTERNAL_CLUSTER_PATH"),
			"https://" + os.Getenv("INTERNAL_CLUSTER_PATH") + ":443",
			"http://localhost",
			"http://localhost:80",
			"https://localhost",
			"https://localhost:443",
			"http://localhost:8082",
			"http://localhost:3000",
			httpOrigin,
		}
		router.Use(cors.New(config))
	} else {
		logger.Info("SetupRouter - CORS disabled")
	}

	// 7. Configure Gin to serve a SPA for non-production env
	// In prod we will use nginx, so this will be ignored!
	// GIN is terrible with SPA, because you can configure static.serve
	// but if you refresh the SPA it will return an error, and you cannot add something like /*
	// The only way is to manage this manually passing the filename in case it's a file, otherwise it must redirect
	// to the index.html page
	if os.Getenv("ENV") != "prod" {
		logger.Info("SetupRouter - Adding NoRoute to handle static files")
		router.NoRoute(func(c *gin.Context) {
			dir, file := path.Split(c.Request.RequestURI)
			ext := filepath.Ext(file)
			allowedExts := []string{".html", ".htm", ".js", ".css", ".json", ".txt", ".jpeg", ".jpg", ".png", ".ico", ".map", ".svg"}
			_, found := utils.Find(allowedExts, ext)
			if found {
				c.File("./public" + path.Join(dir, file))
			} else {
				c.File("./public/index.html")
			}
		})
	} else {
		logger.Info("SetupRouter - Skipping NoRoute config, because it's running in production mode")
	}
	return router, cookieStore
}

// RegisterRoutes function
func RegisterRoutes(ctx context.Context, router *gin.Engine, cookieStore *cookie.Store, logger *zap.SugaredLogger, validate *validator.Validate, client *mongo.Client) {
	oauthGithub = api.NewGithub(ctx, logger, client, oauthCallbackURL, oauthScopes)
	auth = api.NewAuth(ctx, logger)

	keepAlive = api.NewKeepAlive(ctx, logger)
	homes = api.NewHomes(ctx, logger, client, validate)
	devices = api.NewDevices(ctx, logger, client)
	assignDevices = api.NewAssignDevice(ctx, logger, client, validate)
	devicesValues = api.NewDevicesValues(ctx, logger, client, validate)
	profiles = api.NewProfiles(ctx, logger, client, validate)
	register = api.NewRegister(ctx, logger, client, validate)
	// FCM = Firebase Cloud Messaging => identify a smartphone on Firebase to send notifications
	fcmToken = api.NewFCMToken(ctx, logger, client, validate)
	online = api.NewOnline(ctx, logger, client)

	// 1. Configure oAuth2 authentication
	router.Use(sessions.Sessions("session", *cookieStore)) // session called "session"
	// 2. Define public APIs
	// public API to get Login URL
	router.GET("/api/login", oauthGithub.GetLoginURL)
	// public APIs
	router.POST("/api/register", register.PostRegister)
	router.POST("/api/fcmtoken", fcmToken.PostFCMToken)
	router.GET("/api/keepalive", keepAlive.GetKeepAlive)
	// oAuth2 config to register the oauth callback API
	authorized := router.Group("/api/callback")
	authorized.Use(oauthGithub.OauthAuth())
	authorized.GET("", auth.LoginCallback)

	// 3. Define private APIs (/api group) protected via JWTMiddleware
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
		private.POST("/profiles/:id/tokens", profiles.PostProfilesAPIToken)
		private.POST("/profiles/:id/fcmTokens", profiles.PostProfilesFCMToken)

		private.GET("/devices", devices.GetDevices)
		private.PUT("/devices/:id", assignDevices.PutAssignDeviceToHomeRoom)
		private.DELETE("/devices/:id", devices.DeleteDevice)

		private.GET("/devices/:id/values", devicesValues.GetValuesDevice)
		private.POST("/devices/:id/values", devicesValues.PostValueDevice)

		private.GET("/online/:id", online.GetOnline)
	}
}
