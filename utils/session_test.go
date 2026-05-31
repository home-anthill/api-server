package utils

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type fakeSession struct {
	values map[interface{}]interface{}
}

func (s fakeSession) ID() string                                 { return "" }
func (s fakeSession) Get(key interface{}) interface{}            { return s.values[key] }
func (s fakeSession) Set(key interface{}, val interface{})       {}
func (s fakeSession) Delete(key interface{})                     {}
func (s fakeSession) Clear()                                     {}
func (s fakeSession) AddFlash(value interface{}, vars ...string) {}
func (s fakeSession) Flashes(vars ...string) []interface{}       { return nil }
func (s fakeSession) Options(options sessions.Options)           {}
func (s fakeSession) Save() error                                { return nil }

func TestGetProfileFromSession(t *testing.T) {
	profileID := bson.NewObjectID()

	t.Run("returns primitive identity", func(t *testing.T) {
		got, err := GetProfileFromSession(fakeSession{values: map[interface{}]interface{}{
			"profileID": profileID.Hex(),
			"githubID":  int64(123),
		}})
		if err != nil {
			t.Fatalf("GetProfileFromSession returned error: %v", err)
		}
		if got.ID != profileID || got.GithubID != 123 {
			t.Fatalf("GetProfileFromSession() = %#v", got)
		}
	})

	t.Run("rejects missing profile id", func(t *testing.T) {
		if _, err := GetProfileFromSession(fakeSession{values: map[interface{}]interface{}{}}); err == nil {
			t.Fatal("expected missing profileID error")
		}
	})

	t.Run("rejects invalid profile id", func(t *testing.T) {
		_, err := GetProfileFromSession(fakeSession{values: map[interface{}]interface{}{
			"profileID": "bad",
			"githubID":  int64(123),
		}})
		if err == nil {
			t.Fatal("expected invalid profileID error")
		}
	})

	t.Run("rejects missing github id", func(t *testing.T) {
		_, err := GetProfileFromSession(fakeSession{values: map[interface{}]interface{}{
			"profileID": profileID.Hex(),
		}})
		if err == nil {
			t.Fatal("expected missing githubID error")
		}
	})
}

func TestGetProfileFromContext(t *testing.T) {
	profileID := bson.NewObjectID()

	t.Run("returns identity from JWT claims", func(t *testing.T) {
		c := testGinContext()
		c.Set("jwt_claims", &JWTClaims{ProfileID: profileID.Hex(), ID: 123})

		got, err := GetProfileFromContext(c)
		if err != nil {
			t.Fatalf("GetProfileFromContext returned error: %v", err)
		}
		if got.ID != profileID || got.GithubID != 123 {
			t.Fatalf("GetProfileFromContext() = %#v", got)
		}
	})

	t.Run("rejects missing claims", func(t *testing.T) {
		if _, err := GetProfileFromContext(testGinContext()); err == nil {
			t.Fatal("expected missing claims error")
		}
	})

	t.Run("rejects wrong claims type", func(t *testing.T) {
		c := testGinContext()
		c.Set("jwt_claims", "bad")
		if _, err := GetProfileFromContext(c); err == nil {
			t.Fatal("expected wrong claims type error")
		}
	})

	t.Run("rejects invalid profile id", func(t *testing.T) {
		c := testGinContext()
		c.Set("jwt_claims", &JWTClaims{ProfileID: "bad", ID: 123})
		if _, err := GetProfileFromContext(c); err == nil {
			t.Fatal("expected invalid profile id error")
		}
	})

	t.Run("rejects missing github id", func(t *testing.T) {
		c := testGinContext()
		c.Set("jwt_claims", &JWTClaims{ProfileID: profileID.Hex()})
		if _, err := GetProfileFromContext(c); err == nil {
			t.Fatal("expected missing github id error")
		}
	})
}

func TestGetLoggedProfileFromContext(t *testing.T) {
	t.Run("returns claims error before querying database", func(t *testing.T) {
		if _, err := GetLoggedProfileFromContext(testGinContext(), nil); err == nil {
			t.Fatal("expected claims error")
		}
	})

	t.Run("returns database error when profile cannot be loaded", func(t *testing.T) {
		client, err := mongo.Connect(options.Client().
			ApplyURI("mongodb://127.0.0.1:1").
			SetServerSelectionTimeout(10 * time.Millisecond))
		if err != nil {
			t.Fatalf("mongo.Connect returned error: %v", err)
		}
		defer func() { _ = client.Disconnect(context.Background()) }()

		profileID := bson.NewObjectID()
		c := testGinContext()
		c.Request = httptest.NewRequest("GET", "/", nil)
		c.Set("jwt_claims", &JWTClaims{ProfileID: profileID.Hex(), ID: 123})

		_, err = GetLoggedProfileFromContext(c, client.Database("unit").Collection("profiles"))
		if err == nil {
			t.Fatal("expected database error")
		}
	})
}

func TestContains(t *testing.T) {
	id := bson.NewObjectID()
	if !Contains([]bson.ObjectID{bson.NewObjectID(), id}, id) {
		t.Fatal("Contains returned false for present ObjectID")
	}
	if Contains([]bson.ObjectID{bson.NewObjectID()}, id) {
		t.Fatal("Contains returned true for missing ObjectID")
	}
}

func testGinContext() *gin.Context {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)
	return c
}
