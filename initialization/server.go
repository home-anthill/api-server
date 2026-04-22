package initialization

import (
	"api-server/api"
	authpkg "api-server/auth"
	"api-server/utils"
	"crypto/sha256"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	limits "github.com/gin-contrib/size"
	"github.com/gin-gonic/contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.uber.org/zap"
)

// SetupRouter function
func SetupRouter(logger *zap.SugaredLogger) *gin.Engine {
	port := os.Getenv("HTTP_PORT")
	httpServer := os.Getenv("HTTP_SERVER")
	httpOrigin := httpServer + ":" + port
	logger.Infof("SetupRouter - httpOrigin = %s", httpOrigin)

	// 2. init session
	// - secretKey signs the session cookie, preventing tampering.
	// - blockKey encrypts the session cookie, preventing the client from reading its contents.
	//	 Without blockKey, the cookie would still be integrity-protected,
	//	 but its contents would be readable by the browser/user.
	secretKey := os.Getenv("COOKIE_SECRET")
	blockKey := sha256.Sum256([]byte(secretKey))
	store := cookie.NewStore([]byte(secretKey), blockKey[:])
	store.Options(sessions.Options{
		Path: "/",
		// The session stores short-lived OAuth state/PKCE data and the minimal profile identity used by JWTMiddleware.
		// Keep its lifetime (MaxAge) aligned with the web accessToken, so a valid access token does not fail
		// because the session expired first.
		MaxAge:   int(authpkg.WebTokenTTL.Seconds()),
		HttpOnly: true,
		Secure:   os.Getenv("ENV") == "prod",
		SameSite: http.SameSiteLaxMode,
	})

	// 3. init GIN
	// Use gin.New() instead of gin.Default() to avoid Gin's built-in Logger middleware,
	// which logs full request details (including bodies that may contain credentials).
	// Recovery() is kept to handle panics gracefully.
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(sessions.Sessions(utils.SessionName, store))
	router.Use(gzip.Gzip(gzip.DefaultCompression))

	// 5. fix a max POST payload size
	const maxRequestBodySize = 1 * 1024 * 1024 // 1 MB
	logger.Info("SetupRouter - set max POST payload size")
	router.Use(limits.RequestSizeLimiter(maxRequestBodySize))

	// 6. Configure CORS
	// - No origin allowed by default
	// - GET,POST, PUT, HEAD methods
	// - Credentials share disabled
	// - Preflight requests cached for 12 hours
	if os.Getenv("HTTP_CORS") == "true" {
		logger.Warnf("SetupRouter - CORS enabled and httpOrigin is = %s", httpOrigin)
		config := cors.DefaultConfig()
		config.AllowCredentials = true
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
			_, file := path.Split(c.Request.URL.Path)
			ext := filepath.Ext(file)
			allowedExts := []string{".html", ".htm", ".js", ".css", ".json", ".txt", ".jpeg", ".jpg", ".png", ".ico", ".map", ".svg"}
			_, found := utils.Find(allowedExts, ext)
			if found {
				// Strip the leading separator so filepath.Join doesn't treat it as absolute
				trimmed := strings.TrimLeft(filepath.FromSlash(c.Request.URL.Path), string(filepath.Separator))
				cleanPath := filepath.Clean(filepath.Join("public", trimmed))
				if !strings.HasPrefix(cleanPath, "public"+string(filepath.Separator)) {
					c.Status(http.StatusBadRequest)
					return
				}
				c.File("./" + cleanPath)
			} else {
				c.File("./public/index.html")
			}
		})
	} else {
		logger.Info("SetupRouter - Skipping NoRoute config, because it's running in production mode")
	}
	return router
}

// RegisterRoutes function
func RegisterRoutes(router *gin.Engine, logger *zap.SugaredLogger, validate *validator.Validate, client *mongo.Client) {
	auth := authpkg.NewAuth(logger, client)

	oauthGithub := api.NewGitHubOAuth(auth, logger, client, "oauth2_state",
		"oauth2_web_pkce_verifier")

	oauthAppGithub := api.NewGitHubOAuthApp(auth, logger, client, "oauth2_app_state",
		"oauth2_app_pkce_challenge")
	oauthCommon := api.NewOAuthCommon(logger, client)

	keepAlive := api.NewKeepAlive(logger)
	homes := api.NewHomes(logger, client, validate)
	devices := api.NewDevices(logger, client, validate)
	devicesValues := api.NewDevicesValues(logger, client, validate)
	profiles := api.NewProfiles(logger, client, validate)
	// FCM = Firebase Cloud Messaging => identify a smartphone on Firebase to send notifications
	fcmToken := api.NewFCMToken(logger, client, validate)
	online := api.NewOnline(logger, client)

	router.GET("/api/keepalive", keepAlive.GetKeepAlive)
	oauth := router.Group("/api/oauth")
	{
		// web app
		oauth.GET("/login", oauthGithub.GitHubLogin)
		oauth.GET("/callback", oauthGithub.GitHubCallback)
		// mobile app
		oauth.GET("/app/login", oauthAppGithub.GitHubAppLogin)
		oauth.GET("/app/callback", oauthAppGithub.GitHubAppCallback)
		oauth.POST("/app/exchange-code", oauthAppGithub.ExchangeAppCode)
		oauth.POST("/app/refresh", oauthCommon.RefreshAppToken)
		// common
		oauth.POST("/refresh", oauthCommon.RefreshToken)
		oauth.POST("/logout", oauthCommon.Logout)
	}

	// Define private APIs (/api group) protected via JWTMiddleware
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
		private.PUT("/devices/:id", devices.PutAssignDeviceToHomeRoom)
		private.DELETE("/devices/:id", devices.DeleteDevice)

		private.GET("/devices/:id/values", devicesValues.GetValuesDevice)
		private.POST("/devices/:id/values", devicesValues.PostValuesDevice)

		private.POST("/fcmtoken", fcmToken.PostFCMToken)
		private.GET("/online/:id", online.GetOnline)
	}
}
