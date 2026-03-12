package testuutils

import (
	"api-server/models"
	"context"
	"time"

	"github.com/onsi/gomega"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func DropAllCollections(ctx context.Context, collProfiles, collHomes, collDevices *mongo.Collection) {
	var err error
	err = collProfiles.Drop(ctx)
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
	err = collHomes.Drop(ctx)
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
	err = collDevices.Drop(ctx)
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
}

func FindAll[T interface{}](ctx context.Context, collection *mongo.Collection) ([]T, error) {
	cur, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return []T{}, err
	}
	defer cur.Close(ctx)
	result := make([]T, 0)
	for cur.Next(ctx) {
		var res T
		cur.Decode(&res)
		result = append(result, res)
	}
	return result, nil
}

func FindOneById[T interface{}](ctx context.Context, collection *mongo.Collection, id bson.ObjectID) (T, error) {
	var model T
	err := collection.FindOne(ctx, bson.M{
		"_id": id,
	}).Decode(&model)
	return model, err
}

func InsertOne(ctx context.Context, collection *mongo.Collection, obj interface{}) error {
	_, err := collection.InsertOne(ctx, obj)
	return err
}

func AssignHomeToProfile(ctx context.Context, collectionProfiles *mongo.Collection, profileId bson.ObjectID, homeId bson.ObjectID) error {
	_, err := collectionProfiles.UpdateOne(
		ctx,
		bson.M{"_id": profileId},
		bson.M{"$push": bson.M{"homes": homeId}},
	)
	return err
}

func AssignDeviceToProfile(ctx context.Context, collectionProfiles *mongo.Collection, profileId bson.ObjectID, deviceId bson.ObjectID) error {
	_, err := collectionProfiles.UpdateOne(
		ctx,
		bson.M{"_id": profileId},
		bson.M{"$push": bson.M{"devices": deviceId}},
	)
	return err
}

func SetAPITokenToProfile(ctx context.Context, collectionProfiles *mongo.Collection, profileId bson.ObjectID, apiToken string) error {
	_, err := collectionProfiles.UpdateOne(
		ctx,
		bson.M{"_id": profileId},
		bson.M{"$set": bson.M{"apiToken": apiToken}},
	)
	return err
}

// AssignDeviceToHomeAndRoom roomId must be inside home with homeId
// This is an unsafe method used only in testing environment bypassing many checks
func AssignDeviceToHomeAndRoom(ctx context.Context, collectionHomes *mongo.Collection, homeId bson.ObjectID, roomId bson.ObjectID, deviceId bson.ObjectID) error {
	var home models.Home
	err := collectionHomes.FindOne(ctx, bson.M{
		"_id": homeId,
	}).Decode(&home)
	if err != nil {
		return err
	}

	filterHome := bson.D{bson.E{Key: "_id", Value: homeId}}
	arrayFiltersRoom := bson.A{bson.M{"x._id": roomId}}
	opts := []options.Lister[options.UpdateOneOptions]{
		options.UpdateOne().SetArrayFilters(arrayFiltersRoom),
	}
	update := bson.M{
		"$push": bson.M{
			"rooms.$[x].devices": deviceId,
		},
		"$set": bson.M{
			"rooms.$[x].modifiedAt": time.Now(),
		},
	}
	_, err = collectionHomes.UpdateOne(ctx, filterHome, update, opts...)
	return err
}
