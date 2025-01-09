package api

import (
	"api-server/db"
	"api-server/models"
	"api-server/utils"
	"context"
	"encoding/gob"
	"errors"
	"fmt"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
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

// LoginGitHub struct
type LoginGitHub struct {
	client             *mongo.Client
	collProfiles       *mongo.Collection
	ctx                context.Context
	logger             *zap.SugaredLogger
	oauthConfig        *oauth2.Config
	sessionStateName   string
	loginPathStateName string
}

// NewLoginGithub function
func NewLoginGithub(ctx context.Context, logger *zap.SugaredLogger, client *mongo.Client, sessionStateName, clientId, clientSecret, redirectURL string, scopes []string) *LoginGitHub {
	gob.Register(models.Profile{})
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
		ctx:                ctx,
		logger:             logger,
		oauthConfig:        oauthConfig,
		sessionStateName:   sessionStateName,
		loginPathStateName: "state",
	}
}

// GetLoginURL function
func (handler *LoginGitHub) GetLoginURL(c *gin.Context) {
	handler.logger.Info("REST - GET - GetLoginURL called")

	// generate a random state code
	// https://medium.com/keycloak/the-importance-of-the-state-parameter-in-oauth-5419c94bef4c
	// The “state” parameter is sent during the initial Authorization Request and sent back from
	// the Authorization Server to the Client along with the Code (that can be later exchanged to a token).
	// The Client should use the content of this parameter to make sure the Code it received matches
	// the Authorization Request it sent.
	// More info here: https://auth0.com/docs/secure/attack-protection/state-parameters
	stateB64, err := utils.RandToken()
	if err != nil {
		handler.logger.Error("REST - GET - GetLoginURL - cannot create a random session token")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unexpected session error"})
		return
	}
	session := sessions.Default(c)
	session.Set(handler.sessionStateName, stateB64)
	err = session.Save()
	if err != nil {
		handler.logger.Error("REST - GET - GetLoginURL - cannot save session")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unexpected error when saving session"})
		return
	}

	// build loginURL adding the random state as query parameter
	loginURL := handler.oauthConfig.AuthCodeURL(stateB64)
	handler.logger.Debug("REST - GET - GetLoginURL - loginURL: ", loginURL)

	noUnicodeString := strings.ReplaceAll(loginURL, "\\u0026", "&amp;")
	handler.logger.Debug("REST - GET - GetLoginURL - result noUnicodeString: ", noUnicodeString)
	c.Redirect(http.StatusFound, noUnicodeString)
}

// OauthAuth function
func (handler *LoginGitHub) OauthAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// verify if the query param state code matches to one in session
		// https://medium.com/keycloak/the-importance-of-the-state-parameter-in-oauth-5419c94bef4c
		// The “state” parameter is sent during the initial Authorization Request and sent back from
		// the Authorization Server to the Client along with the Code (that can be later exchanged to a token).
		// The Client should use the content of this parameter to make sure the Code it received matches
		// the Authorization Request it sent.
		session := sessions.Default(c)
		sessionStateB64 := session.Get(handler.sessionStateName)
		if sessionStateB64 != c.Query(handler.loginPathStateName) {
			handler.logger.Error("OauthAuth - invalid session state: %s ", sessionStateB64)
			c.AbortWithError(http.StatusInternalServerError, fmt.Errorf("invalid session state: %s", sessionStateB64))
			return
		}

		// read current profile from session.
		// if available save it in the context
		if dbProfile, ok := session.Get("profile").(models.Profile); ok {
			// profile is already on session, so you can simply proceed
			handler.logger.Debug("OauthAuth - dbProfile already in session: ", dbProfile)
			c.Set("profile", dbProfile)
			c.Next()
			return
		}

		var dbGithubUser models.GitHub
		if os.Getenv("ENV") != "testing" {
			// read the "code"
			tok, err := handler.oauthConfig.Exchange(context.TODO(), c.Query("code"))
			if err != nil {
				handler.logger.Errorf("OauthAuth - failed to do exchange: %v", err)
				c.AbortWithError(http.StatusBadRequest, fmt.Errorf("failed to do exchange: %v", err))
				return
			}

			// create a new GitHub API client to perform authentication
			client := github.NewClient(handler.oauthConfig.Client(context.TODO(), tok))
			githubClientUser, _, err := client.Users.Get(context.TODO(), "")
			if err != nil {
				handler.logger.Errorf("OauthAuth - failed to get user: %v", err)
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
		if singleUserLoginEmail != "" && dbGithubUser.Email != singleUserLoginEmail {
			handler.logger.Error("OauthAuth - SINGLE_USER_LOGIN_EMAIL is defined, so user with email = " + dbGithubUser.Email + " cannot log in")
			c.AbortWithError(http.StatusForbidden, fmt.Errorf("user with email %s not admitted to this server", dbGithubUser.Email))
			return
		}

		// find profile searching by GitHub id
		var profileFound models.Profile
		err := handler.collProfiles.FindOne(c, bson.M{
			"github.id": dbGithubUser.ID,
		}).Decode(&profileFound)

		if err == nil {
			// profile found
			c.Set("profile", profileFound)
			// populate cookie
			session.Set("profile", profileFound)
			if errSet := session.Save(); errSet != nil {
				handler.logger.Errorf("OauthAuth - failed to save profile in session: %v", errSet)
				c.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to save profile in session %v", errSet))
			}
		} else {
			// there is an error
			if errors.Is(err, mongo.ErrNoDocuments) {
				handler.logger.Debug("OauthAuth - profile not found, creating a new one")
				currentDate := time.Now()
				// profile not found, so create a new profile
				var newProfile models.Profile
				newProfile.ID = primitive.NewObjectID()
				newProfile.Github = dbGithubUser
				newProfile.APIToken = uuid.NewString()
				newProfile.Homes = []primitive.ObjectID{}   // empty slice of ObjectIDs
				newProfile.Devices = []primitive.ObjectID{} // empty slice of ObjectIDs
				newProfile.CreatedAt = currentDate
				newProfile.ModifiedAt = currentDate

				c.Set("profile", newProfile)

				// populate cookie
				session.Set("profile", newProfile)
				if errSave := session.Save(); errSave != nil {
					handler.logger.Errorf("OauthAuth - failed to save profile in session: %v", errSave)
					c.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to save profile in session %v", errSave))
					return
				}

				// add profile to db
				_, errInsProfile := handler.collProfiles.InsertOne(c, newProfile)
				if errInsProfile != nil {
					handler.logger.Errorf("OauthAuth - cannot save new profile on db: %v", errInsProfile)
					c.AbortWithError(http.StatusInternalServerError, fmt.Errorf("cannot save new profile on db: %v", errInsProfile))
					return
				}

				handler.logger.Debug("OauthAuth - New profile added to db!")
			} else {
				// other error
				handler.logger.Errorf("OauthAuth - cannot find profile on db. Unknown reason: %v", err)
				c.AbortWithError(http.StatusInternalServerError, fmt.Errorf("cannot find profile in db: %v", err))
			}
		}
	}
}
