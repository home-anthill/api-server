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
	"github.com/unrolled/secure"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	"os"
	"path"
	"path/filepath"
	"strings"
)

var OauthGithub *api.Github
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

func SetupRouter(httpOrigin string, logger *zap.SugaredLogger) (*gin.Engine, cookie.Store) {
	// init oauthCallbackURL based on httpOrigin
	oauthCallbackURL = httpOrigin + "/api/callback/"

	// init GIN
	router := gin.Default()
	// init session
	cookieStore := cookie.NewStore([]byte("secret"))
	router.Use(sessions.Sessions("session", cookieStore))
	// apply compression
	router.Use(gzip.Gzip(gzip.DefaultCompression))

	// apply security config to GIN
	logger.Info("SetupRouter - starting SECURE middleware...")
	secureMiddleware := secure.New(getSecureOptions(httpOrigin))
	router.Use(func() gin.HandlerFunc {
		return func(c *gin.Context) {
			errSecure := secureMiddleware.Process(c.Writer, c.Request)
			// If there was an error, do not continue.
			if errSecure != nil {
				c.Abort()
				return
			}
			// Avoid header rewrite if response is a redirection.
			if status := c.Writer.Status(); status > 300 && status < 399 {
				c.Abort()
			}
		}
	}())

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

func RegisterRoutes(router *gin.Engine, cookieStore *cookie.Store, ctx context.Context, logger *zap.SugaredLogger, validate *validator.Validate, collProfiles, collHomes, collDevices *mongo.Collection) {
	OauthGithub = api.NewGithub(ctx, logger, collProfiles, oauthCallbackURL, oauthScopes)
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
	router.GET("/api/login", OauthGithub.GetLoginURL)
	// public APIs
	router.POST("/api/register", register.PostRegister)
	router.GET("/api/keepalive", keepAlive.GetKeepAlive)
	// oAuth2 config to register the oauth callback API
	authorized := router.Group("/api/callback")
	authorized.Use(OauthGithub.OauthAuth())
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

func getSecureOptions(httpOrigin string) secure.Options {
	return secure.Options{
		// AllowedHosts is a list of fully qualified domain names that are allowed. Default is empty list, which allows any and all host names.
		//AllowedHosts: []string{
		//  // TODO find a way to use this feature without breaking everything with docker-compose
		//  // It requires a little bit of investigation
		//  httpOrigin,
		//},
		//// AllowedHostsAreRegex determines, if the provided AllowedHosts slice contains valid regular expressions. Default is false.
		//AllowedHostsAreRegex: false,
		//// HostsProxyHeaders is a set of header keys that may hold a proxied hostname value for the request.
		//HostsProxyHeaders: []string{"X-Forwarded-Hosts"},
		//// If SSLRedirect is set to true, then only allow HTTPS requests. Default is false.
		//SSLRedirect: true,
		//// If SSLTemporaryRedirect is true, the a 302 will be used while redirecting. Default is false (301).
		//SSLTemporaryRedirect: false,
		//// SSLHost is the host name that is used to redirect HTTP requests to HTTPS. Default is "", which indicates to use the same host.
		//SSLHost: "ssl.example.com",
		//// SSLHostFunc is a function pointer, the return value of the function is the host name that has same functionality as `SSHost`. Default is nil. If SSLHostFunc is nil, the `SSLHost` option will be used.
		//SSLHostFunc: nil,
		//// SSLProxyHeaders is set of header keys with associated values that would indicate a valid HTTPS request. Useful when using Nginx: `map[string]string{"X-Forwarded-Proto": "https"}`. Default is blank map.
		//SSLProxyHeaders: map[string]string{"X-Forwarded-Proto": "https"},
		//// STSSeconds is the max-age of the Strict-Transport-Security header. Default is 0, which would NOT include the header.
		//STSSeconds: 31536000,
		//// If STSIncludeSubdomains is set to true, the `includeSubdomains` will be appended to the Strict-Transport-Security header. Default is false.
		//STSIncludeSubdomains: true,
		//// If STSPreload is set to true, the `preload` flag will be appended to the Strict-Transport-Security header. Default is false.
		//STSPreload: true,
		//// STS header is only included when the connection is HTTPS. If you want to force it to always be added, set to true. `IsDevelopment` still overrides this. Default is false.
		//ForceSTSHeader: false,
		// If FrameDeny is set to true, adds the X-Frame-Options header with the value of `DENY`. Default is false.
		// forbids a page from being displayed in a frame
		FrameDeny: true,
		// If ContentTypeNosniff is true, adds the X-Content-Type-Options header with the value `nosniff`. Default is false.
		// used to indicate that the MIME types in the Content-Type headers should be followed and not be changed
		ContentTypeNosniff: true,
		//// If BrowserXssFilter is true, adds the X-XSS-Protection header with the value `1; mode=block`. Default is false.
		//// The HTTP X-XSS-Protection response header is a feature of Internet Explorer, Chrome and Safari that stops pages
		//// from loading when they detect reflected cross-site scripting (XSS) attacks.
		//// These protections are largely unnecessary in modern browsers when sites implement a strong Content-Security-Policy
		//// that disables the use of inline JavaScript ('unsafe-inline'). (https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-XSS-Protection)
		// BrowserXssFilter: true,
		// ContentSecurityPolicy allows the Content-Security-Policy header value to be set with a custom value.
		// Default is "". Passing a template string will replace `$NONCE` with a dynamic nonce value of 16 bytes for each request which can be later retrieved using the Nonce function.
		ContentSecurityPolicy: getCsp(),
		// ReferrerPolicy allows the Referrer-Policy header with the value to be set with a custom value. Default is "".
		ReferrerPolicy: "no-referrer",
		// PermissionsPolicy allows the Permissions-Policy header with the value to be set with a custom value. Default is "".
		PermissionsPolicy: "",
		// This will cause the AllowedHosts, SSLRedirect, and STSSeconds/STSIncludeSubdomains options to be ignored during development. When deploying to production, be sure to set this to false.
		IsDevelopment: os.Getenv("ENV") != "prod",
	}
}

func getCsp() string {
	styles := []string{
		"https://*.googleapis.com/",
	}
	fonts := []string{
		"https://*.gstatic.com/",
	}
	images := []string{
		"data:",
		"https://*.google.com/",
		"https://*.googleusercontent.com/",
		"https://*.fbsbx.com/",
		"https://*.gstatic.com/",
		"https://*.githubusercontent.com/",
	}
	connect := []string{
		"https://*.google.com/",
		"https://*.googleusercontent.com/",
		"https://*.fbsbx.com/",
		"https://*.googleapis.com/",
		"https://*.gstatic.com/",
	}
	// allow-popups is required to open urls in other tabs with target _black via javascript
	sandboxes := []string{
		"allow-forms",
		"allow-scripts",
		"allow-same-origin",
		"allow-popups",
	}
	workers := []string{
		"https://*.google.com/",
		"https://*.googleusercontent.com/",
		"https://*.fbsbx.com/",
		"https://*.googleapis.com/",
		"https://*.gstatic.com/",
	}
	// deprecated but still used in older browsers, defines the valid sources
	// for web workers and nested browsing contexts loaded using elements such as <frame> and <iframe>
	childSrc := "child-src 'none'"
	// restricts the URLs which can be loaded using script interfaces. The APIs that are restricted are:
	// <a> ping, Fetch, XMLHttpRequest, WebSocket, EventSource
	connectSrc := "connect-src 'self' " + strings.Join(connect[:], " ")
	// serves as a fallback for the other CSP fetch directives. For more info check:
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Security-Policy/default-src
	defaultSrc := "default-src 'self'"
	// valid sources for fonts loaded using @font-face
	fontSrc := "font-src 'self' " + strings.Join(fonts[:], " ")
	// restricts the URLs which can be used as the target of a form submissions from a given context
	formAction := "form-action 'self'"
	// specifies valid parents that may embed a page using <frame>, <iframe>, <object>, <embed>, or <applet>
	frameAncestors := "frame-ancestors 'none'"
	// specifies valid sources for nested browsing contexts loading using elements such as <frame> and <iframe>
	frameSrc := "frame-src 'none'"
	// specifies valid sources of images and favicons
	imgSrc := "img-src 'self' " + strings.Join(images[:], " ")
	// specifies which manifest can be applied to the resource.
	manifestSrc := "manifest-src 'self'"
	// specifies valid sources for loading media using the <audio> and <video> elements
	mediaSrc := "media-src 'none'"
	// specifies valid sources for the <object>, <embed>, and <applet> elements
	objectSrc := "object-src 'none'"
	// enables a sandbox for the requested resource similar to the <iframe> sandbox attribute.
	// It applies restrictions to a page's actions including preventing popups, preventing the execution
	// of plugins and scripts, and enforcing a same-origin policy.
	sandbox := "sandbox " + strings.Join(sandboxes[:], " ")
	// specifies valid sources for JavaScript. This includes not only URLs loaded directly into <script>
	scriptSrc := "script-src 'self' 'unsafe-inline'"
	// specifies valid sources for sources for stylesheets.
	styleSrc := "style-src 'self' 'unsafe-inline' " + strings.Join(styles[:], " ")
	// specifies valid sources for Worker, SharedWorker, or ServiceWorker scripts
	workerSrc := "worker-src 'self' " + strings.Join(workers[:], " ")
	// base-uri directive restricts the URLs which can be used in a document's <base> element
	// self = Refers to the origin from which the protected document is being served, including the same URL scheme and port number.
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Security-Policy/base-uri
	baseUri := "base-uri 'self'"

	csp := childSrc + "; " +
		connectSrc + "; " +
		defaultSrc + "; " +
		fontSrc + "; " +
		formAction + "; " +
		frameAncestors + "; " +
		frameSrc + "; " +
		imgSrc + "; " +
		manifestSrc + "; " +
		mediaSrc + "; " +
		objectSrc + "; " +
		sandbox + "; " +
		scriptSrc + "; " +
		styleSrc + "; " +
		workerSrc + "; " +
		baseUri + ";"

	return csp
}
