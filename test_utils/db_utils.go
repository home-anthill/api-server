package test_utils

import (
	"api-server/models"
	"context"
	. "github.com/onsi/gomega"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

func DropAllCollections(ctx context.Context, collProfiles, collHomes, collDevices *mongo.Collection) {
	var err error
	err = collProfiles.Drop(ctx)
	Expect(err).ShouldNot(HaveOccurred())
	err = collHomes.Drop(ctx)
	Expect(err).ShouldNot(HaveOccurred())
	err = collDevices.Drop(ctx)
	Expect(err).ShouldNot(HaveOccurred())
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

func FindOneById[T interface{}](ctx context.Context, collection *mongo.Collection, id primitive.ObjectID) (T, error) {
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

func AssignHomeToProfile(ctx context.Context, collectionProfiles *mongo.Collection, profileId primitive.ObjectID, homeId primitive.ObjectID) error {
	_, err := collectionProfiles.UpdateOne(
		ctx,
		bson.M{"_id": profileId},
		bson.M{"$push": bson.M{"homes": homeId}},
	)
	return err
}

func AssignDeviceToProfile(ctx context.Context, collectionProfiles *mongo.Collection, profileId primitive.ObjectID, deviceId primitive.ObjectID) error {
	_, err := collectionProfiles.UpdateOne(
		ctx,
		bson.M{"_id": profileId},
		bson.M{"$push": bson.M{"devices": deviceId}},
	)
	return err
}

// AssignDeviceToHomeAndRoom roomId must be inside home with homeId
func AssignDeviceToHomeAndRoom(ctx context.Context, collectionHomes *mongo.Collection, homeId primitive.ObjectID, roomId primitive.ObjectID, deviceId primitive.ObjectID) error {
	var home models.Home
	err := collectionHomes.FindOne(ctx, bson.M{
		"_id": homeId,
	}).Decode(&home)
	if err != nil {
		return err
	}

	// update room
	filter := bson.D{primitive.E{Key: "_id", Value: homeId}}
	arrayFilters := options.ArrayFilters{Filters: bson.A{bson.M{"x._id": roomId}}}
	upsert := true
	opts := options.UpdateOptions{
		ArrayFilters: &arrayFilters,
		Upsert:       &upsert,
	}
	update := bson.M{
		"$set": bson.M{
			"rooms.$[x].devices":    []primitive.ObjectID{deviceId},
			"rooms.$[x].modifiedAt": time.Now(),
		},
	}
	_, err = collectionHomes.UpdateOne(ctx, filter, update, &opts)
	return err
}
