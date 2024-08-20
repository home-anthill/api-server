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
var keepAlive *api.KeepAlive

var oauthCallbackURL string
var oauthScopes = []string{"repo"} //https://developer.github.com/v3/oauth/#scopes

// SetupRouter function
func SetupRouter(httpServer string, port string, oauthCallback string, logger *zap.SugaredLogger) (*gin.Engine, cookie.Store) {
	// init oauthCallbackURL based on httpOrigin
	oauthCallbackURL = oauthCallback
	logger.Info("SetupRouter - oauthCallbackURL is = " + oauthCallbackURL)

	httpOrigin := httpServer + ":" + port
	logger.Info("SetupRouter - httpOrigin is = " + httpOrigin)

	// init GIN
	router := gin.Default()
	// init session
	cookieStore := cookie.NewStore([]byte("secret"))
	router.Use(sessions.Sessions("session", cookieStore))
	// apply compression
	router.Use(gzip.Gzip(gzip.DefaultCompression))

	// fix a max POST payload size
	logger.Info("SetupRouter - set mac POST payload size")
	router.Use(limits.RequestSizeLimiter(1024 * 1024))

	// 10. Configure CORS
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

	// 11. Configure Gin to serve a Single Page Application
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
func RegisterRoutes(ctx context.Context, router *gin.Engine, cookieStore *cookie.Store, logger *zap.SugaredLogger, validate *validator.Validate, collProfiles, collHomes, collDevices *mongo.Collection) {
	oauthGithub = api.NewGithub(ctx, logger, collProfiles, oauthCallbackURL, oauthScopes)
	auth = api.NewAuth(ctx, logger)
	homes = api.NewHomes(ctx, logger, collHomes, collProfiles, validate)
	devices = api.NewDevices(ctx, logger, collDevices, collProfiles, collHomes)
	assignDevices = api.NewAssignDevice(ctx, logger, collProfiles, collHomes, validate)
	devicesValues = api.NewDevicesValues(ctx, logger, collDevices, collProfiles, collHomes, validate)
	profiles = api.NewProfiles(ctx, logger, collProfiles)
	register = api.NewRegister(ctx, logger, collDevices, collProfiles, validate)
	keepAlive = api.NewKeepAlive(ctx, logger)

	// 12. Configure oAuth2 authentication
	router.Use(sessions.Sessions("session", *cookieStore)) // session called "session"
	// public API to get Login URL
	router.GET("/api/login", oauthGithub.GetLoginURL)
	// public APIs
	router.POST("/api/register", register.PostRegister)
	router.GET("/api/keepalive", keepAlive.GetKeepAlive)
	// oAuth2 config to register the oauth callback API
	authorized := router.Group("/api/callback")
	authorized.Use(oauthGithub.OauthAuth())
	authorized.GET("", auth.LoginCallback)

	// 13. Define /api group protected via JWTMiddleware
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
		private.PUT("/devices/:id", assignDevices.PutAssignDeviceToHomeRoom)
		private.DELETE("/devices/:id", devices.DeleteDevice)

		private.GET("/devices/:id/values", devicesValues.GetValuesDevice)
		private.POST("/devices/:id/values", devicesValues.PostValueDevice)
	}
}
