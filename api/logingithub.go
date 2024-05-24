package api

import (
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

type Github struct {
	collectionProfiles *mongo.Collection
	ctx                context.Context
	logger             *zap.SugaredLogger
	oauthConfig        *oauth2.Config
}

var DbGithubUserTestmock = models.Github{
	ID:        123456,
	Login:     "Test",
	Name:      "Test Test",
	Email:     "test@test.com",
	AvatarURL: "https://avatars.githubusercontent.com/u/123456?v=4",
}

func NewGithub(ctx context.Context, logger *zap.SugaredLogger, collectionProfiles *mongo.Collection, redirectURL string, scopes []string) *Github {
	gob.Register(models.Profile{})
	// init global configuration with received params
	oauthConfig := &oauth2.Config{
		ClientID:     os.Getenv("OAUTH2_CLIENTID"),
		ClientSecret: os.Getenv("OAUTH2_SECRETID"),
		RedirectURL:  redirectURL,
		Scopes:       scopes,
		Endpoint:     oauth2gh.Endpoint,
	}
	return &Github{
		collectionProfiles: collectionProfiles,
		ctx:                ctx,
		logger:             logger,
		oauthConfig:        oauthConfig,
	}
}

func (handler *Github) GetLoginURL(c *gin.Context) {
	handler.logger.Info("REST - GET - GetLoginURL called")

	state, err := utils.RandToken()
	if err != nil {
		handler.logger.Error("REST - GET - GetLoginURL - cannot create a random session token")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unexpected session error"})
		return
	}

	session := sessions.Default(c)
	session.Set("state", state)
	err = session.Save()
	if err != nil {
		handler.logger.Error("REST - GET - GetLoginURL - cannot save session")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unexpected error when saving session"})
		return
	}

	loginURL := handler.oauthConfig.AuthCodeURL(state)
	noUnicodeString := strings.ReplaceAll(loginURL, "\\u0026", "&amp;")
	handler.logger.Info("REST - GET - GetLoginURL - result noUnicodeString: ", noUnicodeString)

	loginUrlRes := models.LoginUrl{}
	loginUrlRes.LoginURL = noUnicodeString
	c.JSON(http.StatusOK, &loginUrlRes)
}

func (handler *Github) OauthAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// read current profile from session.
		// if available save it in the context
		session := sessions.Default(c)
		if dbProfile, ok := session.Get("profile").(models.Profile); ok {
			// profile is already on session, so you can simply proceed
			handler.logger.Debug("OauthAuth - dbProfile already in session: ", dbProfile)
			c.Set("profile", dbProfile)
			c.Next()
			return
		}

		retrievedState := session.Get("state")
		if retrievedState != c.Query("state") {
			handler.logger.Error("OauthAuth - invalid session state: %s ", retrievedState)
			c.AbortWithError(http.StatusInternalServerError, fmt.Errorf("invalid session state: %s", retrievedState))
			return
		}

		var dbGithubUser models.Github
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

			dbGithubUser = models.Github{
				ID:        *githubClientUser.ID,
				Login:     *githubClientUser.Login,
				Name:      *githubClientUser.Name,
				Email:     *githubClientUser.Email,
				AvatarURL: *githubClientUser.AvatarURL,
			}
		} else {
			dbGithubUser = DbGithubUserTestmock
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

		// find profile searching by github.id == githubClientUser.ID
		var profileFound models.Profile
		err := handler.collectionProfiles.FindOne(c, bson.M{
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
				newProfile.ApiToken = uuid.NewString()
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

				// ad profile to db
				_, errInsProfile := handler.collectionProfiles.InsertOne(c, newProfile)
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
