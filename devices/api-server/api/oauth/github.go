package oauth

import (
  "api-server/models"
  "context"
  "crypto/rand"
  "encoding/base64"
  "encoding/gob"
  "fmt"
  "github.com/gin-gonic/contrib/sessions"
  "github.com/gin-gonic/gin"
  "github.com/golang/glog"
  "github.com/google/go-github/github"
  "github.com/google/uuid"
  "go.mongodb.org/mongo-driver/bson"
  "go.mongodb.org/mongo-driver/bson/primitive"
  "go.mongodb.org/mongo-driver/mongo"
  "go.uber.org/zap"
  "golang.org/x/oauth2"
  oauth2gh "golang.org/x/oauth2/github"
  "net/http"
  "os"
  "strings"
  "time"
)

type Credentials struct {
  ClientID     string `json:"clientid"`
  ClientSecret string `json:"secret"`
}

var conf *oauth2.Config
var state string
var store sessions.CookieStore
var collection *mongo.Collection
var logger *zap.SugaredLogger

func init() {
  gob.Register(models.Profile{})
}

func Setup(redirectURL string, scopes []string, log *zap.SugaredLogger, profilesCollection *mongo.Collection) {
  // init some global vars
  logger = log
  collection = profilesCollection
  store = sessions.NewCookieStore([]byte(os.Getenv("COOKIE_SECRET")))

  // init global configuration with received params
  conf = &oauth2.Config{
    ClientID:     os.Getenv("OAUTH2_CLIENTID"),
    ClientSecret: os.Getenv("OAUTH2_SECRETID"),
    RedirectURL:  redirectURL,
    Scopes:       scopes,
    Endpoint:     oauth2gh.Endpoint,
  }
}

func Session(name string) gin.HandlerFunc {
  return sessions.Sessions(name, store)
}

func GetLoginURL(c *gin.Context) {
  logger.Info("REST - GET - GetLoginURL called")

  state = randToken()
  session := sessions.Default(c)
  session.Set("state", state)
  session.Save()
  loginURL := conf.AuthCodeURL(state)
  noUnicodeString := strings.ReplaceAll(loginURL, "\\u0026", "&amp;")
  logger.Info("GetLoginURL result noUnicodeString: ", noUnicodeString)
  c.JSON(http.StatusOK, gin.H{
    "loginURL": noUnicodeString,
  })
}

func OauthAuth() gin.HandlerFunc {
  return func(ctx *gin.Context) {
    // read current profile from session.
    // if available save it in the context
    session := sessions.Default(ctx)
    if dbProfile, ok := session.Get("profile").(models.Profile); ok {
      logger.Info("***** Already in session **** - dbProfile: ", dbProfile)
      ctx.Set("profile", dbProfile)
      ctx.Next()
      return
    }

    // read state query param from context (URL)
    retrievedState := session.Get("state")
    if retrievedState != ctx.Query("state") {
      ctx.AbortWithError(http.StatusUnauthorized, fmt.Errorf("invalid session state: %s", retrievedState))
      return
    }

    // TODO: oauth2.NoContext -> context.Context from stdlib
    // read the "code"
    tok, err := conf.Exchange(context.TODO(), ctx.Query("code"))
    if err != nil {
      ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("failed to do exchange: %v", err))
      return
    }

    // create a new GitHub API client to perform authentication
    client := github.NewClient(conf.Client(context.TODO(), tok))
    var githubClientUser *github.User
    githubClientUser, _, err = client.Users.Get(context.TODO(), "")
    if err != nil {
      ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("failed to get user: %v", err))
      return
    }

    dbGithubUser := models.Github{
      ID:        *githubClientUser.ID,
      Login:     *githubClientUser.Login,
      Name:      *githubClientUser.Name,
      Email:     *githubClientUser.Email,
      AvatarURL: *githubClientUser.AvatarURL,
    }

    // ATTENTION!!!
    // if SINGLE_USER_LOGIN_EMAIL is defined, only an account with Email equals
    // to the one defined in SINGLE_USER_LOGIN_EMAIL env variable can log in to this server.
    singleUserLoginEmail := os.Getenv("SINGLE_USER_LOGIN_EMAIL")
    fmt.Println("singleUserLoginEmail: " + singleUserLoginEmail)
    if singleUserLoginEmail != "" && dbGithubUser.Email != singleUserLoginEmail {
      logger.Error("SINGLE_USER_LOGIN_EMAIL is defined, so user with email = " + dbGithubUser.Email + " cannot log in")
      ctx.AbortWithError(http.StatusForbidden, fmt.Errorf("user with email %s not admitted to this server", dbGithubUser.Email))
      return
    }

    // find profile searching by github.id == githubClientUser.ID
    var profileFound models.Profile
    err = collection.FindOne(ctx, bson.M{
      "github.id": githubClientUser.ID,
    }).Decode(&profileFound)

    if err == nil {
      // profile found
      ctx.Set("profile", profileFound)
      // populate cookie
      session.Set("profile", profileFound)
      if errSet := session.Save(); errSet != nil {
        glog.Errorf("Failed to save profile in session: %v", errSet)
      }
    } else {
      // there is an error
      if err == mongo.ErrNoDocuments {
        logger.Info("Profile not found, creating a new one...")
        // profile not found, so create a new profile
        var newProfile models.Profile
        newProfile.ID = primitive.NewObjectID()
        newProfile.Github = dbGithubUser
        newProfile.ApiToken = uuid.NewString()
        newProfile.Homes = []primitive.ObjectID{}   // empty slice of ObjectIDs
        newProfile.Devices = []primitive.ObjectID{} // empty slice of ObjectIDs
        newProfile.CreatedAt = time.Now()
        newProfile.ModifiedAt = time.Now()

        ctx.Set("profile", newProfile)

        // populate cookie
        session.Set("profile", newProfile)
        if errSave := session.Save(); errSave != nil {
          glog.Errorf("Failed to save profile in session: %v", errSave)
        }

        // ad profile to db
        _, err2 := collection.InsertOne(ctx, newProfile)
        if err2 != nil {
          ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("cannot save new profile on db: %v", err2))
          return
        }
        logger.Info("New profile added to db!")
      } else {
        // other error
        logger.Error("Cannot find profile on db. Unknown reason: ", err)
        ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("cannot find profile in db: %v", err))
      }
    }
  }
}

func randToken() string {
  b := make([]byte, 32)
  if _, err := rand.Read(b); err != nil {
    glog.Fatalf("[Gin-OAuth] Failed to read rand: %v\n", err)
  }
  return base64.StdEncoding.EncodeToString(b)
}
