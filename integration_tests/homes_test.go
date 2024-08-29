package integration_tests

import (
	"api-server/api"
	"api-server/db"
	"api-server/initialization"
	"api-server/models"
	"api-server/testuutils"
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	"net/http"
	"net/http/httptest"
	"os"
	"time"
)

var _ = Describe("Homes", func() {
	var ctx context.Context
	var logger *zap.SugaredLogger
	var router *gin.Engine
	var client *mongo.Client
	var collProfiles *mongo.Collection
	var collHomes *mongo.Collection
	var collDevices *mongo.Collection

	var currentDate1 = time.Now()
	var currentDate2 = time.Now()
	var home1 = models.Home{
		ID:       primitive.NewObjectID(),
		Name:     "home1",
		Location: "location1",
		Rooms: []models.Room{{
			ID:         primitive.NewObjectID(),
			Name:       "room1",
			Floor:      1,
			CreatedAt:  currentDate1,
			ModifiedAt: currentDate1,
			Devices:    []primitive.ObjectID{},
		}, {
			ID:         primitive.NewObjectID(),
			Name:       "room2",
			Floor:      2,
			CreatedAt:  currentDate1,
			ModifiedAt: currentDate1,
			Devices:    []primitive.ObjectID{},
		}},
		CreatedAt:  currentDate1,
		ModifiedAt: currentDate1,
	}
	var home2 = models.Home{
		ID:         primitive.NewObjectID(),
		Name:       "home2",
		Location:   "location2",
		Rooms:      []models.Room{},
		CreatedAt:  currentDate2,
		ModifiedAt: currentDate2,
	}

	BeforeEach(func() {
		logger, router, ctx, client = initialization.Start()
		defer logger.Sync()

		collProfiles = db.GetCollections(client).Profiles
		collHomes = db.GetCollections(client).Homes
		collDevices = db.GetCollections(client).Devices

		err := os.Setenv("SINGLE_USER_LOGIN_EMAIL", "test@test.com")
		Expect(err).ShouldNot(HaveOccurred())
	})

	AfterEach(func() {
		testuutils.DropAllCollections(ctx, collProfiles, collHomes, collDevices)
	})

	Context("calling homes api GET", func() {
		BeforeEach(func() {
			err := testuutils.InsertOne(ctx, collHomes, home1)
			Expect(err).ShouldNot(HaveOccurred())
			err = testuutils.InsertOne(ctx, collHomes, home2)
			Expect(err).ShouldNot(HaveOccurred())
		})

		When("profile doesn't own any homes", func() {
			It("should get a list of empty homes", func() {
				jwtToken, cookieSession := testuutils.GetJwt(router)
				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodGet, "/api/homes", nil)
				req.Header.Add("Cookie", cookieSession)
				req.Header.Add("Authorization", "Bearer "+jwtToken)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)
				Expect(recorder.Code).To(Equal(http.StatusOK))
				var homes []models.Home
				err := json.Unmarshal(recorder.Body.Bytes(), &homes)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(homes).To(HaveLen(0))
			})
		})

		When("profile owns an home", func() {
			It("should get a list of homes", func() {
				jwtToken, cookieSession := testuutils.GetJwt(router)
				profileRes := testuutils.GetLoggedProfile(router, jwtToken, cookieSession)

				err := testuutils.AssignHomeToProfile(ctx, collProfiles, profileRes.ID, home1.ID)
				Expect(err).ShouldNot(HaveOccurred())

				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodGet, "/api/homes", nil)
				req.Header.Add("Cookie", cookieSession)
				req.Header.Add("Authorization", "Bearer "+jwtToken)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)
				Expect(recorder.Code).To(Equal(http.StatusOK))
				var homes []models.Home
				err = json.Unmarshal(recorder.Body.Bytes(), &homes)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(homes).To(HaveLen(1))
				Expect(homes[0].ID).To(Equal(home1.ID))
				Expect(homes[0].Name).To(Equal(home1.Name))
				Expect(homes[0].Location).To(Equal(home1.Location))
				Expect(homes[0].Rooms).To(HaveLen(len(home1.Rooms)))
				Expect(homes[0].Rooms[0].ID).To(Equal(home1.Rooms[0].ID))
				Expect(homes[0].Rooms[0].Name).To(Equal(home1.Rooms[0].Name))
				Expect(homes[0].Rooms[0].Floor).To(Equal(home1.Rooms[0].Floor))
				//Expect(homes[0].Rooms[0].CreatedAt).To(Equal(home1.Rooms[0].CreatedAt))
				//Expect(homes[0].Rooms[0].ModifiedAt).To(Equal(home1.Rooms[0].ModifiedAt))
				Expect(homes[0].Rooms[0].Devices).To(ContainElements(home1.Rooms[0].Devices))
				Expect(homes[0].Rooms[1].ID).To(Equal(home1.Rooms[1].ID))
				Expect(homes[0].Rooms[1].Name).To(Equal(home1.Rooms[1].Name))
				Expect(homes[0].Rooms[1].Floor).To(Equal(home1.Rooms[1].Floor))
				//Expect(homes[0].Rooms[1].CreatedAt).To(Equal(home1.Rooms[1].CreatedAt))
				//Expect(homes[0].Rooms[1].ModifiedAt).To(Equal(home1.Rooms[1].ModifiedAt))
				Expect(homes[0].Rooms[1].Devices).To(ContainElements(home1.Rooms[1].Devices))
				//Expect(homes[0].CreatedAt).To(Equal(home1.CreatedAt))
				//Expect(homes[0].ModifiedAt).To(Equal(home1.ModifiedAt))
			})
		})

		When("profile owns more homes", func() {
			It("should get a list of homes", func() {
				jwtToken, cookieSession := testuutils.GetJwt(router)
				profileRes := testuutils.GetLoggedProfile(router, jwtToken, cookieSession)

				err := testuutils.AssignHomeToProfile(ctx, collProfiles, profileRes.ID, home1.ID)
				Expect(err).ShouldNot(HaveOccurred())
				err = testuutils.AssignHomeToProfile(ctx, collProfiles, profileRes.ID, home2.ID)
				Expect(err).ShouldNot(HaveOccurred())

				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodGet, "/api/homes", nil)
				req.Header.Add("Cookie", cookieSession)
				req.Header.Add("Authorization", "Bearer "+jwtToken)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)
				Expect(recorder.Code).To(Equal(http.StatusOK))
				var homes []models.Home
				err = json.Unmarshal(recorder.Body.Bytes(), &homes)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(homes).To(HaveLen(2))
				Expect(homes[0].ID).To(Equal(home1.ID))
				Expect(homes[0].Name).To(Equal(home1.Name))
				Expect(homes[0].Location).To(Equal(home1.Location))
				Expect(homes[0].Rooms).To(HaveLen(len(home1.Rooms)))
				Expect(homes[1].ID).To(Equal(home2.ID))
				Expect(homes[1].Name).To(Equal(home2.Name))
				Expect(homes[1].Location).To(Equal(home2.Location))
				Expect(homes[1].Rooms).To(HaveLen(len(home2.Rooms)))
			})
		})
	})

	Context("calling homes api POST", func() {
		When("profile doesn't own any homes", func() {
			It("should create a new home and assign it to the logged profile", func() {
				home3 := api.HomeNewReq{
					Name:     "home3",
					Location: "location3",
					Rooms: []api.RoomNewReq{{
						Name:  "room1",
						Floor: 1,
					}},
				}
				var homeBuf bytes.Buffer
				err := json.NewEncoder(&homeBuf).Encode(home3)
				Expect(err).ShouldNot(HaveOccurred())

				jwtToken, cookieSession := testuutils.GetJwt(router)
				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodPost, "/api/homes", &homeBuf)
				req.Header.Add("Cookie", cookieSession)
				req.Header.Add("Authorization", "Bearer "+jwtToken)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)
				Expect(recorder.Code).To(Equal(http.StatusOK))
				var homeResult models.Home
				err = json.Unmarshal(recorder.Body.Bytes(), &homeResult)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(homeResult.Name).To(Equal(home3.Name))
				Expect(homeResult.Location).To(Equal(home3.Location))
				Expect(homeResult.Rooms[0].Name).To(Equal(home3.Rooms[0].Name))
				Expect(homeResult.Rooms[0].Floor).To(Equal(home3.Rooms[0].Floor))
			})

			It("should return an error, if body is missing", func() {
				jwtToken, cookieSession := testuutils.GetJwt(router)
				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodPost, "/api/homes", nil)
				req.Header.Add("Cookie", cookieSession)
				req.Header.Add("Authorization", "Bearer "+jwtToken)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)
				Expect(recorder.Code).To(Equal(http.StatusBadRequest))
				Expect(recorder.Body.String()).To(Equal(`{"error":"invalid request payload"}`))
			})

			It("should return an error, if body is not valid", func() {
				home3 := api.HomeNewReq{
					Name:     "", // not valid, because length must be > 0
					Location: "", // not valid, because length must be > 0
					Rooms: []api.RoomNewReq{{
						Name:  "",  // not valid, because length must be > 0
						Floor: -55, // not valid, because must be >= -50 and <= 300
					}},
				}
				var homeBuf bytes.Buffer
				err := json.NewEncoder(&homeBuf).Encode(home3)
				Expect(err).ShouldNot(HaveOccurred())

				jwtToken, cookieSession := testuutils.GetJwt(router)
				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodPost, "/api/homes", &homeBuf)
				req.Header.Add("Cookie", cookieSession)
				req.Header.Add("Authorization", "Bearer "+jwtToken)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)
				Expect(recorder.Code).To(Equal(http.StatusBadRequest))
				Expect(recorder.Body.String()).To(Equal(`{"error":"invalid request body, these fields are not valid: name location name floor"}`))
			})
		})
	})

	Context("calling homes api PUT", func() {
		BeforeEach(func() {
			err := testuutils.InsertOne(ctx, collHomes, home1)
			Expect(err).ShouldNot(HaveOccurred())
		})

		When("profile owns a single home", func() {
			It("should update an existing home", func() {
				jwtToken, cookieSession := testuutils.GetJwt(router)
				profileRes := testuutils.GetLoggedProfile(router, jwtToken, cookieSession)

				err := testuutils.AssignHomeToProfile(ctx, collProfiles, profileRes.ID, home1.ID)
				Expect(err).ShouldNot(HaveOccurred())

				updateHome1 := api.HomeUpdateReq{
					Name:     "home3",
					Location: "location3",
				}
				var homeBuf bytes.Buffer
				err = json.NewEncoder(&homeBuf).Encode(updateHome1)
				Expect(err).ShouldNot(HaveOccurred())

				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodPut, "/api/homes/"+home1.ID.Hex(), &homeBuf)
				req.Header.Add("Cookie", cookieSession)
				req.Header.Add("Authorization", "Bearer "+jwtToken)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)
				Expect(recorder.Code).To(Equal(http.StatusOK))
				Expect(recorder.Body.String()).To(Equal(`{"message":"home has been updated"}`))

				home1FromDb, err := testuutils.FindOneById[models.Home](ctx, collHomes, home1.ID)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(home1FromDb.Name).To(Equal(updateHome1.Name))
				Expect(home1FromDb.Location).To(Equal(updateHome1.Location))
				Expect(home1FromDb.Rooms[0].Name).To(Equal(home1.Rooms[0].Name))
				Expect(home1FromDb.Rooms[0].Floor).To(Equal(home1.Rooms[0].Floor))
			})

			It("should return an error, if homeId is wrong", func() {
				jwtToken, cookieSession := testuutils.GetJwt(router)
				recorder := httptest.NewRecorder()
				badHomeID := "bad_home_id"
				req := httptest.NewRequest(http.MethodPut, "/api/homes/"+badHomeID, nil)
				req.Header.Add("Cookie", cookieSession)
				req.Header.Add("Authorization", "Bearer "+jwtToken)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)
				Expect(recorder.Code).To(Equal(http.StatusBadRequest))
				Expect(recorder.Body.String()).To(Equal(`{"error":"wrong format of the path param 'id'"}`))
			})

			It("should return an error, if body is missing", func() {
				jwtToken, cookieSession := testuutils.GetJwt(router)
				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodPut, "/api/homes/"+home1.ID.Hex(), nil)
				req.Header.Add("Cookie", cookieSession)
				req.Header.Add("Authorization", "Bearer "+jwtToken)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)
				Expect(recorder.Code).To(Equal(http.StatusBadRequest))
				Expect(recorder.Body.String()).To(Equal(`{"error":"invalid request payload"}`))
			})

			It("should return an error, if body is not valid", func() {
				updateHome1 := api.HomeUpdateReq{
					Name:     "", // not valid, because length must be > 0
					Location: "", // not valid, because length must be > 0
				}
				var homeBuf bytes.Buffer
				err := json.NewEncoder(&homeBuf).Encode(updateHome1)
				Expect(err).ShouldNot(HaveOccurred())

				jwtToken, cookieSession := testuutils.GetJwt(router)
				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodPut, "/api/homes/"+home1.ID.Hex(), &homeBuf)
				req.Header.Add("Cookie", cookieSession)
				req.Header.Add("Authorization", "Bearer "+jwtToken)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)
				Expect(recorder.Code).To(Equal(http.StatusBadRequest))
				Expect(recorder.Body.String()).To(Equal(`{"error":"invalid request body, these fields are not valid: name location"}`))
			})

			It("should return an error, if profile is not the owner of this home", func() {
				jwtToken, cookieSession := testuutils.GetJwt(router)

				updateHome1 := api.HomeUpdateReq{
					Name:     "home3",
					Location: "location3",
				}
				var homeBuf bytes.Buffer
				err := json.NewEncoder(&homeBuf).Encode(updateHome1)
				Expect(err).ShouldNot(HaveOccurred())

				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodPut, "/api/homes/"+home1.ID.Hex(), &homeBuf)
				req.Header.Add("Cookie", cookieSession)
				req.Header.Add("Authorization", "Bearer "+jwtToken)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)
				Expect(recorder.Code).To(Equal(http.StatusBadRequest))
				Expect(recorder.Body.String()).To(Equal(`{"error":"cannot update a home that is not in your profile"}`))
			})
		})
	})

	Context("calling homes api DELETE", func() {
		BeforeEach(func() {
			err := testuutils.InsertOne(ctx, collHomes, home1)
			Expect(err).ShouldNot(HaveOccurred())
		})

		When("profile owns an home", func() {
			It("should delete a home", func() {
				jwtToken, cookieSession := testuutils.GetJwt(router)
				profileRes := testuutils.GetLoggedProfile(router, jwtToken, cookieSession)

				err := testuutils.AssignHomeToProfile(ctx, collProfiles, profileRes.ID, home1.ID)
				Expect(err).ShouldNot(HaveOccurred())

				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodDelete, "/api/homes/"+home1.ID.Hex(), nil)
				req.Header.Add("Cookie", cookieSession)
				req.Header.Add("Authorization", "Bearer "+jwtToken)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)
				Expect(recorder.Code).To(Equal(http.StatusOK))
				Expect(recorder.Body.String()).To(Equal(`{"message":"home has been deleted"}`))

				_, err = testuutils.FindOneById[models.Home](ctx, collHomes, home1.ID)
				Expect(err).Should(HaveOccurred())
			})

			It("should return an error, if homeId is wrong", func() {
				jwtToken, cookieSession := testuutils.GetJwt(router)
				recorder := httptest.NewRecorder()
				badHomeID := "bad_home_id"
				req := httptest.NewRequest(http.MethodDelete, "/api/homes/"+badHomeID, nil)
				req.Header.Add("Cookie", cookieSession)
				req.Header.Add("Authorization", "Bearer "+jwtToken)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)
				Expect(recorder.Code).To(Equal(http.StatusBadRequest))
				Expect(recorder.Body.String()).To(Equal(`{"error":"wrong format of the path param 'id'"}`))
			})

			It("should return an error, if profile is not the owner of this home", func() {
				jwtToken, cookieSession := testuutils.GetJwt(router)
				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodDelete, "/api/homes/"+home1.ID.Hex(), nil)
				req.Header.Add("Cookie", cookieSession)
				req.Header.Add("Authorization", "Bearer "+jwtToken)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)
				Expect(recorder.Code).To(Equal(http.StatusBadRequest))
				Expect(recorder.Body.String()).To(Equal(`{"error":"cannot delete a home that is not in your profile"}`))
			})
		})

		When("profile owns more home", func() {
			It("should delete a home", func() {
				jwtToken, cookieSession := testuutils.GetJwt(router)
				profileRes := testuutils.GetLoggedProfile(router, jwtToken, cookieSession)

				err := testuutils.InsertOne(ctx, collHomes, home2)
				Expect(err).ShouldNot(HaveOccurred())
				err = testuutils.AssignHomeToProfile(ctx, collProfiles, profileRes.ID, home1.ID)
				Expect(err).ShouldNot(HaveOccurred())
				err = testuutils.AssignHomeToProfile(ctx, collProfiles, profileRes.ID, home2.ID)
				Expect(err).ShouldNot(HaveOccurred())

				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodDelete, "/api/homes/"+home1.ID.Hex(), nil)
				req.Header.Add("Cookie", cookieSession)
				req.Header.Add("Authorization", "Bearer "+jwtToken)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)
				Expect(recorder.Code).To(Equal(http.StatusOK))
				Expect(recorder.Body.String()).To(Equal(`{"message":"home has been deleted"}`))

				// `home1` should be removed via delete api
				_, err = testuutils.FindOneById[models.Home](ctx, collHomes, home1.ID)
				Expect(err).Should(HaveOccurred())
				// `home2` should be still in db
				homeFound, err := testuutils.FindOneById[models.Home](ctx, collHomes, home2.ID)
				Expect(err).Should(Not(HaveOccurred()))
				Expect(homeFound.ID).To(Equal(home2.ID))
			})
		})
	})

	Context("calling rooms api GET", func() {
		BeforeEach(func() {
			err := testuutils.InsertOne(ctx, collHomes, home1)
			Expect(err).ShouldNot(HaveOccurred())
		})

		When("profile owns an home with rooms", func() {
			It("should get the list of rooms of an home", func() {
				jwtToken, cookieSession := testuutils.GetJwt(router)
				profileRes := testuutils.GetLoggedProfile(router, jwtToken, cookieSession)

				err := testuutils.AssignHomeToProfile(ctx, collProfiles, profileRes.ID, home1.ID)
				Expect(err).ShouldNot(HaveOccurred())

				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodGet, "/api/homes/"+home1.ID.Hex()+"/rooms", nil)
				req.Header.Add("Cookie", cookieSession)
				req.Header.Add("Authorization", "Bearer "+jwtToken)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)
				Expect(recorder.Code).To(Equal(http.StatusOK))
				var rooms []models.Room
				err = json.Unmarshal(recorder.Body.Bytes(), &rooms)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(rooms).To(HaveLen(2))
				Expect(rooms[0].Name).To(Equal(home1.Rooms[0].Name))
				Expect(rooms[0].Floor).To(Equal(home1.Rooms[0].Floor))
				Expect(rooms[1].Name).To(Equal(home1.Rooms[1].Name))
				Expect(rooms[1].Floor).To(Equal(home1.Rooms[1].Floor))
			})

			It("should return an error, if homeId is wrong", func() {
				jwtToken, cookieSession := testuutils.GetJwt(router)
				badHomeID := "bad_home_id"
				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodGet, "/api/homes/"+badHomeID+"/rooms", nil)
				req.Header.Add("Cookie", cookieSession)
				req.Header.Add("Authorization", "Bearer "+jwtToken)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)
				Expect(recorder.Code).To(Equal(http.StatusBadRequest))
				Expect(recorder.Body.String()).To(Equal(`{"error":"wrong format of the path param 'id'"}`))
			})

			It("should return an error, if profile is not the owner of this home", func() {
				jwtToken, cookieSession := testuutils.GetJwt(router)
				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodGet, "/api/homes/"+home1.ID.Hex()+"/rooms", nil)
				req.Header.Add("Cookie", cookieSession)
				req.Header.Add("Authorization", "Bearer "+jwtToken)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)
				Expect(recorder.Code).To(Equal(http.StatusBadRequest))
				Expect(recorder.Body.String()).To(Equal(`{"error":"cannot get rooms of an home that is not in your profile"}`))
			})

			It("should return an error, if homeId is not in db", func() {
				jwtToken, cookieSession := testuutils.GetJwt(router)
				profileRes := testuutils.GetLoggedProfile(router, jwtToken, cookieSession)

				missingHomeID := primitive.NewObjectID()

				err := testuutils.AssignHomeToProfile(ctx, collProfiles, profileRes.ID, missingHomeID)
				Expect(err).ShouldNot(HaveOccurred())

				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodGet, "/api/homes/"+missingHomeID.Hex()+"/rooms", nil)
				req.Header.Add("Cookie", cookieSession)
				req.Header.Add("Authorization", "Bearer "+jwtToken)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)
				Expect(recorder.Code).To(Equal(http.StatusNotFound))
				Expect(recorder.Body.String()).To(Equal(`{"error":"cannot find rooms for that home"}`))
			})
		})
	})

	Context("calling rooms api POST", func() {
		BeforeEach(func() {
			err := testuutils.InsertOne(ctx, collHomes, home2)
			Expect(err).ShouldNot(HaveOccurred())
		})

		When("profile owns an home", func() {
			It("should add a new room to the home", func() {
				room1 := api.RoomNewReq{
					Name:  "room1",
					Floor: 2,
				}
				var roomBuf bytes.Buffer
				err := json.NewEncoder(&roomBuf).Encode(room1)
				Expect(err).ShouldNot(HaveOccurred())

				jwtToken, cookieSession := testuutils.GetJwt(router)
				profileRes := testuutils.GetLoggedProfile(router, jwtToken, cookieSession)

				err = testuutils.AssignHomeToProfile(ctx, collProfiles, profileRes.ID, home2.ID)
				Expect(err).ShouldNot(HaveOccurred())

				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodPost, "/api/homes/"+home2.ID.Hex()+"/rooms", &roomBuf)
				req.Header.Add("Cookie", cookieSession)
				req.Header.Add("Authorization", "Bearer "+jwtToken)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)
				Expect(recorder.Code).To(Equal(http.StatusOK))
				Expect(recorder.Body.String()).To(Equal(`{"message":"room added to the home"}`))

				home2FromDb, err := testuutils.FindOneById[models.Home](ctx, collHomes, home2.ID)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(home2FromDb.Rooms).To(HaveLen(1))
				Expect(home2FromDb.Rooms[0].Name).To(Equal(room1.Name))
				Expect(home2FromDb.Rooms[0].Floor).To(Equal(room1.Floor))
			})
		})

		When("receive bad inputs", func() {
			It("should return an error, because homeId is wrong", func() {
				jwtToken, cookieSession := testuutils.GetJwt(router)
				badHomeID := "bad_home_id"
				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodPost, "/api/homes/"+badHomeID+"/rooms", nil)
				req.Header.Add("Cookie", cookieSession)
				req.Header.Add("Authorization", "Bearer "+jwtToken)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)
				Expect(recorder.Code).To(Equal(http.StatusBadRequest))
				Expect(recorder.Body.String()).To(Equal(`{"error":"wrong format of the path param 'id'"}`))
			})

			It("should return an error, because body is missing", func() {
				jwtToken, cookieSession := testuutils.GetJwt(router)
				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodPost, "/api/homes/"+home1.ID.Hex()+"/rooms", nil)
				req.Header.Add("Cookie", cookieSession)
				req.Header.Add("Authorization", "Bearer "+jwtToken)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)
				Expect(recorder.Code).To(Equal(http.StatusBadRequest))
				Expect(recorder.Body.String()).To(Equal(`{"error":"invalid request payload"}`))
			})

			It("should return an error, because body is not valid", func() {
				room1 := api.RoomNewReq{
					Name:  "",  // not valid, because length must be > 0
					Floor: -55, // not valid, because it must be >= -50 and <= 300
				}
				var roomBuf bytes.Buffer
				err := json.NewEncoder(&roomBuf).Encode(room1)
				Expect(err).ShouldNot(HaveOccurred())

				jwtToken, cookieSession := testuutils.GetJwt(router)
				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodPost, "/api/homes/"+home1.ID.Hex()+"/rooms", &roomBuf)
				req.Header.Add("Cookie", cookieSession)
				req.Header.Add("Authorization", "Bearer "+jwtToken)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)
				Expect(recorder.Code).To(Equal(http.StatusBadRequest))
				Expect(recorder.Body.String()).To(Equal(`{"error":"invalid request body, these fields are not valid: name floor"}`))
			})

			It("should return an error, because room of home is not owned by profile", func() {
				room1 := api.RoomNewReq{
					Name:  "room1",
					Floor: 2,
				}
				var roomBuf bytes.Buffer
				err := json.NewEncoder(&roomBuf).Encode(room1)
				Expect(err).ShouldNot(HaveOccurred())

				jwtToken, cookieSession := testuutils.GetJwt(router)
				// profile doesn't have `home2` assigned to it
				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodPost, "/api/homes/"+home2.ID.Hex()+"/rooms", &roomBuf)
				req.Header.Add("Cookie", cookieSession)
				req.Header.Add("Authorization", "Bearer "+jwtToken)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)
				Expect(recorder.Code).To(Equal(http.StatusBadRequest))
				Expect(recorder.Body.String()).To(Equal(`{"error":"cannot create a room in an home that is not in your profile"}`))
			})

			It("should return an error, because homeId is not in db", func() {
				room1 := api.RoomNewReq{
					Name:  "room1",
					Floor: 2,
				}
				var roomBuf bytes.Buffer
				err := json.NewEncoder(&roomBuf).Encode(room1)
				Expect(err).ShouldNot(HaveOccurred())

				jwtToken, cookieSession := testuutils.GetJwt(router)
				profileRes := testuutils.GetLoggedProfile(router, jwtToken, cookieSession)
				missingHomeID := primitive.NewObjectID()

				err = testuutils.AssignHomeToProfile(ctx, collProfiles, profileRes.ID, missingHomeID)
				Expect(err).ShouldNot(HaveOccurred())

				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodPost, "/api/homes/"+missingHomeID.Hex()+"/rooms", &roomBuf)
				req.Header.Add("Cookie", cookieSession)
				req.Header.Add("Authorization", "Bearer "+jwtToken)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)
				Expect(recorder.Code).To(Equal(http.StatusNotFound))
				Expect(recorder.Body.String()).To(Equal(`{"error":"cannot find home"}`))
			})
		})
	})

	Context("calling rooms api PUT", func() {
		BeforeEach(func() {
			err := testuutils.InsertOne(ctx, collHomes, home1)
			Expect(err).ShouldNot(HaveOccurred())
		})

		When("profile owns an home", func() {
			It("should update an existing room of the home", func() {
				room1Upd := api.RoomUpdateReq{
					Name:  "room1-upd",
					Floor: 0,
				}
				var roomBuf bytes.Buffer
				err := json.NewEncoder(&roomBuf).Encode(room1Upd)
				Expect(err).ShouldNot(HaveOccurred())

				jwtToken, cookieSession := testuutils.GetJwt(router)
				profileRes := testuutils.GetLoggedProfile(router, jwtToken, cookieSession)

				err = testuutils.AssignHomeToProfile(ctx, collProfiles, profileRes.ID, home1.ID)
				Expect(err).ShouldNot(HaveOccurred())

				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodPut, "/api/homes/"+home1.ID.Hex()+"/rooms/"+home1.Rooms[0].ID.Hex(), &roomBuf)
				req.Header.Add("Cookie", cookieSession)
				req.Header.Add("Authorization", "Bearer "+jwtToken)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)
				Expect(recorder.Code).To(Equal(http.StatusOK))
				Expect(recorder.Body.String()).To(Equal(`{"message":"room has been updated"}`))

				home1FromDb, err := testuutils.FindOneById[models.Home](ctx, collHomes, home1.ID)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(home1FromDb.Rooms).To(HaveLen(len(home1.Rooms)))
				Expect(home1FromDb.Rooms[0].Name).To(Equal(room1Upd.Name))
				Expect(home1FromDb.Rooms[0].Floor).To(Equal(room1Upd.Floor))
			})
		})

		When("receive bad inputs", func() {
			It("should return an error, because homeId is wrong", func() {
				jwtToken, cookieSession := testuutils.GetJwt(router)
				badHomeID := "bad_home_id"
				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodPut, "/api/homes/"+badHomeID+"/rooms/"+home1.Rooms[0].ID.Hex(), nil)
				req.Header.Add("Cookie", cookieSession)
				req.Header.Add("Authorization", "Bearer "+jwtToken)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)
				Expect(recorder.Code).To(Equal(http.StatusBadRequest))
				Expect(recorder.Body.String()).To(Equal(`{"error":"wrong format of one of the path params"}`))
			})

			It("should return an error, because roomId is wrong", func() {
				jwtToken, cookieSession := testuutils.GetJwt(router)
				badRoomID := "bad_room_id"
				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodPut, "/api/homes/"+home1.ID.Hex()+"/rooms/"+badRoomID, nil)
				req.Header.Add("Cookie", cookieSession)
				req.Header.Add("Authorization", "Bearer "+jwtToken)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)
				Expect(recorder.Code).To(Equal(http.StatusBadRequest))
				Expect(recorder.Body.String()).To(Equal(`{"error":"wrong format of one of the path params"}`))
			})

			It("should return an error, because body is missing", func() {
				jwtToken, cookieSession := testuutils.GetJwt(router)
				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodPut, "/api/homes/"+home1.ID.Hex()+"/rooms/"+home1.Rooms[0].ID.Hex(), nil)
				req.Header.Add("Cookie", cookieSession)
				req.Header.Add("Authorization", "Bearer "+jwtToken)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)
				Expect(recorder.Code).To(Equal(http.StatusBadRequest))
				Expect(recorder.Body.String()).To(Equal(`{"error":"invalid request payload"}`))
			})

			It("should return an error, because body is not valid", func() {
				room1Upd := api.RoomUpdateReq{
					Name:  "",  // not valid, because length must be > 0
					Floor: -55, // not valid, because it must be >= -50 and <= 300
				}
				var roomBuf bytes.Buffer
				err := json.NewEncoder(&roomBuf).Encode(room1Upd)
				Expect(err).ShouldNot(HaveOccurred())

				jwtToken, cookieSession := testuutils.GetJwt(router)
				//profileRes := test_utils.GetLoggedProfile(router, jwtToken, cookieSession)
				//
				//err = test_utils.AssignHomeToProfile(ctx, collProfiles, profileRes.ID, home1.ID)
				//Expect(err).ShouldNot(HaveOccurred())

				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodPut, "/api/homes/"+home1.ID.Hex()+"/rooms/"+home1.Rooms[0].ID.Hex(), &roomBuf)
				req.Header.Add("Cookie", cookieSession)
				req.Header.Add("Authorization", "Bearer "+jwtToken)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)
				Expect(recorder.Code).To(Equal(http.StatusBadRequest))
				Expect(recorder.Body.String()).To(Equal(`{"error":"invalid request body, these fields are not valid: name floor"}`))
			})

			It("should return an error, because room of home is not owned by profile", func() {
				room1Upd := api.RoomUpdateReq{
					Name:  "room1-upd",
					Floor: 0,
				}
				var roomBuf bytes.Buffer
				err := json.NewEncoder(&roomBuf).Encode(room1Upd)
				Expect(err).ShouldNot(HaveOccurred())

				jwtToken, cookieSession := testuutils.GetJwt(router)
				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodPut, "/api/homes/"+home1.ID.Hex()+"/rooms/"+home1.Rooms[0].ID.Hex(), &roomBuf)
				req.Header.Add("Cookie", cookieSession)
				req.Header.Add("Authorization", "Bearer "+jwtToken)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)
				Expect(recorder.Code).To(Equal(http.StatusBadRequest))
				Expect(recorder.Body.String()).To(Equal(`{"error":"cannot update a room in an home that is not in your profile"}`))
			})

			It("should return an error, because homeId is not in db", func() {
				room1Upd := api.RoomUpdateReq{
					Name:  "room1-upd",
					Floor: 0,
				}
				var roomBuf bytes.Buffer
				err := json.NewEncoder(&roomBuf).Encode(room1Upd)
				Expect(err).ShouldNot(HaveOccurred())

				jwtToken, cookieSession := testuutils.GetJwt(router)
				profileRes := testuutils.GetLoggedProfile(router, jwtToken, cookieSession)
				missingHomeID := primitive.NewObjectID()

				err = testuutils.AssignHomeToProfile(ctx, collProfiles, profileRes.ID, missingHomeID)
				Expect(err).ShouldNot(HaveOccurred())

				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodPut, "/api/homes/"+missingHomeID.Hex()+"/rooms/"+home1.Rooms[0].ID.Hex(), &roomBuf)
				req.Header.Add("Cookie", cookieSession)
				req.Header.Add("Authorization", "Bearer "+jwtToken)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)
				Expect(recorder.Code).To(Equal(http.StatusNotFound))
				Expect(recorder.Body.String()).To(Equal(`{"error":"cannot find rooms for that home"}`))
			})

			It("should return an error, because roomId is not a room of home1 in db", func() {
				room1Upd := api.RoomUpdateReq{
					Name:  "room1-upd",
					Floor: 0,
				}
				var roomBuf bytes.Buffer
				err := json.NewEncoder(&roomBuf).Encode(room1Upd)
				Expect(err).ShouldNot(HaveOccurred())

				jwtToken, cookieSession := testuutils.GetJwt(router)
				profileRes := testuutils.GetLoggedProfile(router, jwtToken, cookieSession)
				missingRoomID := primitive.NewObjectID()

				err = testuutils.AssignHomeToProfile(ctx, collProfiles, profileRes.ID, home1.ID)
				Expect(err).ShouldNot(HaveOccurred())

				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodPut, "/api/homes/"+home1.ID.Hex()+"/rooms/"+missingRoomID.Hex(), &roomBuf)
				req.Header.Add("Cookie", cookieSession)
				req.Header.Add("Authorization", "Bearer "+jwtToken)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)
				Expect(recorder.Code).To(Equal(http.StatusNotFound))
				Expect(recorder.Body.String()).To(Equal(`{"error":"room not found"}`))
			})
		})
	})

	Context("calling rooms api DELETE", func() {
		BeforeEach(func() {
			err := testuutils.InsertOne(ctx, collHomes, home1)
			Expect(err).ShouldNot(HaveOccurred())
		})

		When("profile owns an home", func() {
			It("should delete a room of that home", func() {
				jwtToken, cookieSession := testuutils.GetJwt(router)
				profileRes := testuutils.GetLoggedProfile(router, jwtToken, cookieSession)

				err := testuutils.AssignHomeToProfile(ctx, collProfiles, profileRes.ID, home1.ID)
				Expect(err).ShouldNot(HaveOccurred())

				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodDelete, "/api/homes/"+home1.ID.Hex()+"/rooms/"+home1.Rooms[0].ID.Hex(), nil)
				req.Header.Add("Cookie", cookieSession)
				req.Header.Add("Authorization", "Bearer "+jwtToken)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)
				Expect(recorder.Code).To(Equal(http.StatusOK))
				Expect(recorder.Body.String()).To(Equal(`{"message":"room has been deleted"}`))

				home1FromDb, err := testuutils.FindOneById[models.Home](ctx, collHomes, home1.ID)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(home1FromDb.Rooms).To(HaveLen(len(home1.Rooms) - 1))
				Expect(home1FromDb.Rooms[0].Name).To(Equal(home1.Rooms[1].Name))
				Expect(home1FromDb.Rooms[0].Floor).To(Equal(home1.Rooms[1].Floor))
			})
		})

		When("receive bad inputs", func() {
			It("should return an error, because homeId is wrong", func() {
				jwtToken, cookieSession := testuutils.GetJwt(router)
				badHomeID := "bad_home_id"
				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodDelete, "/api/homes/"+badHomeID+"/rooms/"+home1.Rooms[0].ID.Hex(), nil)
				req.Header.Add("Cookie", cookieSession)
				req.Header.Add("Authorization", "Bearer "+jwtToken)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)
				Expect(recorder.Code).To(Equal(http.StatusBadRequest))
				Expect(recorder.Body.String()).To(Equal(`{"error":"wrong format of one of the path params"}`))
			})

			It("should return an error, because roomId is wrong", func() {
				jwtToken, cookieSession := testuutils.GetJwt(router)
				badRoomID := "bad_room_id"
				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodDelete, "/api/homes/"+home1.ID.Hex()+"/rooms/"+badRoomID, nil)
				req.Header.Add("Cookie", cookieSession)
				req.Header.Add("Authorization", "Bearer "+jwtToken)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)
				Expect(recorder.Code).To(Equal(http.StatusBadRequest))
				Expect(recorder.Body.String()).To(Equal(`{"error":"wrong format of one of the path params"}`))
			})

			It("should return an error, because home is not owned by profile", func() {
				jwtToken, cookieSession := testuutils.GetJwt(router)
				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodDelete, "/api/homes/"+home1.ID.Hex()+"/rooms/"+home1.Rooms[0].ID.Hex(), nil)
				req.Header.Add("Cookie", cookieSession)
				req.Header.Add("Authorization", "Bearer "+jwtToken)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)
				Expect(recorder.Code).To(Equal(http.StatusBadRequest))
				Expect(recorder.Body.String()).To(Equal(`{"error":"cannot delete a room in an home that is not in your profile"}`))
			})

			It("should return an error, because homeId is not in db", func() {
				jwtToken, cookieSession := testuutils.GetJwt(router)
				profileRes := testuutils.GetLoggedProfile(router, jwtToken, cookieSession)

				missingHomeID := primitive.NewObjectID()
				err := testuutils.AssignHomeToProfile(ctx, collProfiles, profileRes.ID, missingHomeID)
				Expect(err).ShouldNot(HaveOccurred())

				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodDelete, "/api/homes/"+missingHomeID.Hex()+"/rooms/"+home1.Rooms[0].ID.Hex(), nil)
				req.Header.Add("Cookie", cookieSession)
				req.Header.Add("Authorization", "Bearer "+jwtToken)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)
				Expect(recorder.Code).To(Equal(http.StatusNotFound))
				Expect(recorder.Body.String()).To(Equal(`{"error":"home not found"}`))
			})

			It("should return an error, because roomId is not a room of home1 in db", func() {
				jwtToken, cookieSession := testuutils.GetJwt(router)
				profileRes := testuutils.GetLoggedProfile(router, jwtToken, cookieSession)
				missingRoomID := primitive.NewObjectID()

				err := testuutils.AssignHomeToProfile(ctx, collProfiles, profileRes.ID, home1.ID)
				Expect(err).ShouldNot(HaveOccurred())

				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodDelete, "/api/homes/"+home1.ID.Hex()+"/rooms/"+missingRoomID.Hex(), nil)
				req.Header.Add("Cookie", cookieSession)
				req.Header.Add("Authorization", "Bearer "+jwtToken)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)
				Expect(recorder.Code).To(Equal(http.StatusNotFound))
				Expect(recorder.Body.String()).To(Equal(`{"error":"room not found"}`))
			})
		})
	})
})
