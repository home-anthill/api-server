package integration_tests

import (
	"api-server/api"
	"api-server/init_config"
	"api-server/models"
	"api-server/test_utils"
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
		// 1. Init config
		logger = init_config.BuildConfig()
		defer logger.Sync()

		err := os.Setenv("SINGLE_USER_LOGIN_EMAIL", "test@test.com")
		Expect(err).ShouldNot(HaveOccurred())

		// 2. Init server
		port := os.Getenv("HTTP_PORT")
		httpOrigin := os.Getenv("HTTP_SERVER") + ":" + port

		router, ctx, collProfiles, collHomes, collDevices = init_config.BuildServer(httpOrigin, logger)
	})

	AfterEach(func() {
		test_utils.DropAllCollections(ctx, collProfiles, collHomes, collDevices)
	})

	Context("calling homes api GET", func() {
		BeforeEach(func() {
			err := test_utils.InsertOne(ctx, collHomes, home1)
			Expect(err).ShouldNot(HaveOccurred())
			err = test_utils.InsertOne(ctx, collHomes, home2)
			Expect(err).ShouldNot(HaveOccurred())
		})

		When("profile doesn't own any homes", func() {
			It("should get a list of empty homes", func() {
				jwtToken, cookieSession := test_utils.GetJwt(router)
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
				jwtToken, cookieSession := test_utils.GetJwt(router)
				profileRes := test_utils.GetLoggedProfile(router, jwtToken, cookieSession)

				err := test_utils.AssignHomeToProfile(ctx, collProfiles, profileRes.ID, home1.ID)
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

				jwtToken, cookieSession := test_utils.GetJwt(router)
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
		})
	})

	Context("calling homes api PUT", func() {
		BeforeEach(func() {
			err := test_utils.InsertOne(ctx, collHomes, home1)
			Expect(err).ShouldNot(HaveOccurred())
		})

		When("profile owns a single home", func() {
			It("should update an existing home", func() {
				jwtToken, cookieSession := test_utils.GetJwt(router)
				profileRes := test_utils.GetLoggedProfile(router, jwtToken, cookieSession)

				err := test_utils.AssignHomeToProfile(ctx, collProfiles, profileRes.ID, home1.ID)
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
				Expect(recorder.Body.String()).To(Equal(`{"message":"Home has been updated"}`))

				home1FromDb, err := test_utils.FindOneById[models.Home](ctx, collHomes, home1.ID)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(home1FromDb.Name).To(Equal(updateHome1.Name))
				Expect(home1FromDb.Location).To(Equal(updateHome1.Location))
				Expect(home1FromDb.Rooms[0].Name).To(Equal(home1.Rooms[0].Name))
				Expect(home1FromDb.Rooms[0].Floor).To(Equal(home1.Rooms[0].Floor))
			})
		})
	})

	Context("calling homes api DELETE", func() {
		BeforeEach(func() {
			err := test_utils.InsertOne(ctx, collHomes, home1)
			Expect(err).ShouldNot(HaveOccurred())
		})

		When("profile owns an home", func() {
			It("should get a list of homes", func() {
				jwtToken, cookieSession := test_utils.GetJwt(router)
				profileRes := test_utils.GetLoggedProfile(router, jwtToken, cookieSession)

				err := test_utils.AssignHomeToProfile(ctx, collProfiles, profileRes.ID, home1.ID)
				Expect(err).ShouldNot(HaveOccurred())

				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodDelete, "/api/homes/"+home1.ID.Hex(), nil)
				req.Header.Add("Cookie", cookieSession)
				req.Header.Add("Authorization", "Bearer "+jwtToken)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)
				Expect(recorder.Code).To(Equal(http.StatusOK))
				Expect(recorder.Body.String()).To(Equal(`{"message":"Home has been deleted"}`))

				_, err = test_utils.FindOneById[models.Home](ctx, collHomes, home1.ID)
				Expect(err).Should(HaveOccurred())
			})
		})
	})

	Context("calling rooms api GET", func() {
		BeforeEach(func() {
			err := test_utils.InsertOne(ctx, collHomes, home1)
			Expect(err).ShouldNot(HaveOccurred())
		})

		When("profile owns an home with rooms", func() {
			It("should get the list of rooms of an home", func() {
				jwtToken, cookieSession := test_utils.GetJwt(router)
				profileRes := test_utils.GetLoggedProfile(router, jwtToken, cookieSession)

				err := test_utils.AssignHomeToProfile(ctx, collProfiles, profileRes.ID, home1.ID)
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
		})
	})

	Context("calling rooms api POST", func() {
		BeforeEach(func() {
			err := test_utils.InsertOne(ctx, collHomes, home2)
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

				jwtToken, cookieSession := test_utils.GetJwt(router)
				profileRes := test_utils.GetLoggedProfile(router, jwtToken, cookieSession)

				err = test_utils.AssignHomeToProfile(ctx, collProfiles, profileRes.ID, home2.ID)
				Expect(err).ShouldNot(HaveOccurred())

				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodPost, "/api/homes/"+home2.ID.Hex()+"/rooms", &roomBuf)
				req.Header.Add("Cookie", cookieSession)
				req.Header.Add("Authorization", "Bearer "+jwtToken)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)
				Expect(recorder.Code).To(Equal(http.StatusOK))
				Expect(recorder.Body.String()).To(Equal(`{"message":"Room added to the home"}`))

				home2FromDb, err := test_utils.FindOneById[models.Home](ctx, collHomes, home2.ID)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(home2FromDb.Rooms).To(HaveLen(1))
				Expect(home2FromDb.Rooms[0].Name).To(Equal(room1.Name))
				Expect(home2FromDb.Rooms[0].Floor).To(Equal(room1.Floor))
			})
		})
	})

	Context("calling rooms api PUT", func() {
		BeforeEach(func() {
			err := test_utils.InsertOne(ctx, collHomes, home1)
			Expect(err).ShouldNot(HaveOccurred())
		})

		When("profile owns an home", func() {
			It("should update an existing room of the home", func() {
				room1Upd := api.RoomUpdateReq{
					Name:    "room1-upd",
					Floor:   1,
					Devices: []primitive.ObjectID{},
				}
				var roomBuf bytes.Buffer
				err := json.NewEncoder(&roomBuf).Encode(room1Upd)
				Expect(err).ShouldNot(HaveOccurred())

				jwtToken, cookieSession := test_utils.GetJwt(router)
				profileRes := test_utils.GetLoggedProfile(router, jwtToken, cookieSession)

				err = test_utils.AssignHomeToProfile(ctx, collProfiles, profileRes.ID, home1.ID)
				Expect(err).ShouldNot(HaveOccurred())

				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodPut, "/api/homes/"+home1.ID.Hex()+"/rooms/"+home1.Rooms[0].ID.Hex(), &roomBuf)
				req.Header.Add("Cookie", cookieSession)
				req.Header.Add("Authorization", "Bearer "+jwtToken)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)
				Expect(recorder.Code).To(Equal(http.StatusOK))
				Expect(recorder.Body.String()).To(Equal(`{"message":"Room has been updated"}`))

				home1FromDb, err := test_utils.FindOneById[models.Home](ctx, collHomes, home1.ID)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(home1FromDb.Rooms).To(HaveLen(len(home1.Rooms)))
				Expect(home1FromDb.Rooms[0].Name).To(Equal(room1Upd.Name))
				Expect(home1FromDb.Rooms[0].Floor).To(Equal(room1Upd.Floor))
			})
		})
	})

	Context("calling rooms api DELETE", func() {
		BeforeEach(func() {
			err := test_utils.InsertOne(ctx, collHomes, home1)
			Expect(err).ShouldNot(HaveOccurred())
		})

		When("profile owns an home", func() {
			It("should delete a room of that home", func() {
				jwtToken, cookieSession := test_utils.GetJwt(router)
				profileRes := test_utils.GetLoggedProfile(router, jwtToken, cookieSession)

				err := test_utils.AssignHomeToProfile(ctx, collProfiles, profileRes.ID, home1.ID)
				Expect(err).ShouldNot(HaveOccurred())

				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodDelete, "/api/homes/"+home1.ID.Hex()+"/rooms/"+home1.Rooms[0].ID.Hex(), nil)
				req.Header.Add("Cookie", cookieSession)
				req.Header.Add("Authorization", "Bearer "+jwtToken)
				req.Header.Add("Content-Type", `application/json`)
				router.ServeHTTP(recorder, req)
				Expect(recorder.Code).To(Equal(http.StatusOK))
				Expect(recorder.Body.String()).To(Equal(`{"message":"Room has been deleted"}`))

				home1FromDb, err := test_utils.FindOneById[models.Home](ctx, collHomes, home1.ID)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(home1FromDb.Rooms).To(HaveLen(len(home1.Rooms) - 1))
				Expect(home1FromDb.Rooms[0].Name).To(Equal(home1.Rooms[1].Name))
				Expect(home1FromDb.Rooms[0].Floor).To(Equal(home1.Rooms[1].Floor))
			})
		})
	})

})
