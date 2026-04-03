package api

import (
	"api-server/db"
	"api-server/models"
	"api-server/utils"
	"crypto/subtle"
	"encoding/gob"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/google/go-github/github"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	oauth2gh "golang.org/x/oauth2/github"
)

// LoginGitHub handles GitHub OAuth2 login and profile creation.
type LoginGitHub struct {
	client             *mongo.Client
	collProfiles       *mongo.Collection
	logger             *zap.SugaredLogger
	oauthConfig        *oauth2.Config
	sessionStateName   string
	loginPathStateName string
}

func init() {
	gob.Register(models.Profile{})
}

// NewLoginGithub constructs a LoginGitHub handler with the given OAuth2 credentials.
func NewLoginGithub(logger *zap.SugaredLogger, client *mongo.Client, sessionStateName, clientId, clientSecret, redirectURL string, scopes []string) *LoginGitHub {
	// init global configuration with received params
	oauthConfig := &oauth2.Config{
		ClientID:     clientId,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       scopes,
		Endpoint:     oauth2gh.Endpoint,
	}
	return &LoginGitHub{
		client:             client,
		collProfiles:       db.GetCollections(client).Profiles,
		logger:             logger,
		oauthConfig:        oauthConfig,
		sessionStateName:   sessionStateName,
		loginPathStateName: "state",
	}
}

// GetLoginURL function
func (lg *LoginGitHub) GetLoginURL(c *gin.Context) {
	lg.logger.Info("REST - GET - GetLoginURL called")

	// generate a random state code
	// https://medium.com/keycloak/the-importance-of-the-state-parameter-in-oauth-5419c94bef4c
	// The “state” parameter is sent during the initial Authorization Request and sent back from
	// the Authorization Server to the Client along with the Code (that can be later exchanged to a token).
	// The Client should use the content of this parameter to make sure the Code it received matches
	// the Authorization Request it sent.
	// More info here: https://auth0.com/docs/secure/attack-protection/state-parameters
	stateB64, err := utils.RandToken()
	if err != nil {
		lg.logger.Error("REST - GET - GetLoginURL - cannot create a random session token")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unexpected session error"})
		return
	}
	session := sessions.Default(c)
	session.Set(lg.sessionStateName, stateB64)
	err = session.Save()
	if err != nil {
		lg.logger.Error("REST - GET - GetLoginURL - cannot save session")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unexpected error when saving session"})
		return
	}

	// build loginURL adding the random state as query parameter
	loginURL := lg.oauthConfig.AuthCodeURL(stateB64)
	lg.logger.Debug("REST - GET - GetLoginURL - loginURL: ", loginURL)

	noUnicodeString := strings.ReplaceAll(loginURL, "\\u0026", "&amp;")
	lg.logger.Debug("REST - GET - GetLoginURL - result noUnicodeString: ", noUnicodeString)
	c.Redirect(http.StatusFound, noUnicodeString)
}

// OauthAuth function
func (lg *LoginGitHub) OauthAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// verify if the query param state code matches to one in session
		// https://medium.com/keycloak/the-importance-of-the-state-parameter-in-oauth-5419c94bef4c
		// The “state” parameter is sent during the initial Authorization Request and sent back from
		// the Authorization Server to the Client along with the Code (that can be later exchanged to a token).
		// The Client should use the content of this parameter to make sure the Code it received matches
		// the Authorization Request it sent.
		session := sessions.Default(c)
		sessionStateB64 := session.Get(lg.sessionStateName)
		sessionState, ok := sessionStateB64.(string)
		if !ok || sessionState == "" {
			lg.logger.Error("OauthAuth - missing or invalid session state")
			c.AbortWithError(http.StatusInternalServerError, fmt.Errorf("invalid session state"))
			return
		}
		if subtle.ConstantTimeCompare([]byte(sessionState), []byte(c.Query(lg.loginPathStateName))) != 1 {
			lg.logger.Error("OauthAuth - invalid session state")
			c.AbortWithError(http.StatusInternalServerError, fmt.Errorf("invalid session state"))
			return
		}

		// read current profile from session.
		// if available save it in the context
		if dbProfile, ok := session.Get("profile").(models.Profile); ok {
			// profile is already on session, so you can simply proceed
			lg.logger.Debug("OauthAuth - dbProfile already in session: ", dbProfile)
			c.Set("profile", dbProfile)
			c.Next()
			return
		}

		var dbGithubUser models.GitHub
		if os.Getenv("ENV") != "testing" {
			// read the "code"
			tok, err := lg.oauthConfig.Exchange(c.Request.Context(), c.Query("code"))
			if err != nil {
				lg.logger.Errorf("OauthAuth - failed to do exchange: %v", err)
				c.AbortWithError(http.StatusBadRequest, fmt.Errorf("failed to do exchange: %v", err))
				return
			}

			// create a new GitHub API client to perform authentication
			client := github.NewClient(lg.oauthConfig.Client(c.Request.Context(), tok))
			githubClientUser, _, err := client.Users.Get(c.Request.Context(), "")
			if err != nil {
				lg.logger.Errorf("OauthAuth - failed to get user: %v", err)
				c.AbortWithError(http.StatusBadRequest, fmt.Errorf("failed to get user: %v", err))
				return
			}

			dbGithubUser = models.GitHub{
				ID:        *githubClientUser.ID,
				Login:     *githubClientUser.Login,
				Name:      *githubClientUser.Name,
				Email:     *githubClientUser.Email,
				AvatarURL: *githubClientUser.AvatarURL,
			}
		} else {
			dbGithubUser = models.DbGithubUserTestmock
		}

		// ATTENTION!!!
		// if SINGLE_USER_LOGIN_EMAIL is defined, only an account with Email equals
		// to the one defined in SINGLE_USER_LOGIN_EMAIL env variable can log in to this server.
		singleUserLoginEmail := os.Getenv("SINGLE_USER_LOGIN_EMAIL")
		if singleUserLoginEmail != "" {
			if dbGithubUser.Email == "" {
				lg.logger.Error("OauthAuth - SINGLE_USER_LOGIN_EMAIL is defined but GitHub user has no email")
				c.AbortWithError(http.StatusForbidden, fmt.Errorf("login not permitted"))
				return
			}
			if dbGithubUser.Email != singleUserLoginEmail {
				lg.logger.Errorf("OauthAuth - SINGLE_USER_LOGIN_EMAIL is defined, so user with email = %s cannot log in", dbGithubUser.Email)
				c.AbortWithError(http.StatusForbidden, fmt.Errorf("login not permitted"))
				return
			}
		}

		// find profile searching by GitHub id
		var profileFound models.Profile
		err := lg.collProfiles.FindOne(c.Request.Context(), bson.M{
			"github.id": dbGithubUser.ID,
		}).Decode(&profileFound)

		if err == nil {
			// profile found
			c.Set("profile", profileFound)
			// populate cookie
			session.Set("profile", profileFound)
			if errSet := session.Save(); errSet != nil {
				lg.logger.Errorf("OauthAuth - failed to save profile in session: %v", errSet)
				c.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to save profile in session %v", errSet))
			}
			lg.logger.Infow("AUDIT - user login",
				"profileID", profileFound.ID.Hex(),
				"githubLogin", dbGithubUser.Login,
				"clientIP", c.ClientIP(),
			)
		} else {
			// there is an error
			if errors.Is(err, mongo.ErrNoDocuments) {
				lg.logger.Debug("OauthAuth - profile not found, creating a new one")
				currentDate := time.Now()
				// profile not found, so create a new profile
				var newProfile models.Profile
				newProfile.ID = bson.NewObjectID()
				newProfile.Github = dbGithubUser
				newProfile.APIToken = uuid.NewString()
				newProfile.Homes = []bson.ObjectID{}   // empty slice of ObjectIDs
				newProfile.Devices = []bson.ObjectID{} // empty slice of ObjectIDs
				newProfile.CreatedAt = currentDate
				newProfile.ModifiedAt = currentDate

				c.Set("profile", newProfile)

				// populate cookie
				session.Set("profile", newProfile)
				if errSave := session.Save(); errSave != nil {
					lg.logger.Errorf("OauthAuth - failed to save profile in session: %v", errSave)
					c.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to save profile in session %v", errSave))
					return
				}

				// add profile to db
				_, errInsProfile := lg.collProfiles.InsertOne(c.Request.Context(), newProfile)
				if errInsProfile != nil {
					lg.logger.Errorf("OauthAuth - cannot save new profile on db: %v", errInsProfile)
					c.AbortWithError(http.StatusInternalServerError, fmt.Errorf("cannot save new profile on db: %v", errInsProfile))
					return
				}

				lg.logger.Debug("OauthAuth - New profile added to db!")
				lg.logger.Infow("AUDIT - user created",
					"profileID", newProfile.ID.Hex(),
					"githubLogin", dbGithubUser.Login,
					"clientIP", c.ClientIP(),
				)
			} else {
				// other error
				lg.logger.Errorf("OauthAuth - cannot find profile on db. Unknown reason: %v", err)
				c.AbortWithError(http.StatusInternalServerError, fmt.Errorf("cannot find profile in db: %v", err))
			}
		}
	}
}
