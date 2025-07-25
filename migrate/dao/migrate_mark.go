package foundationdao

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	metaerror "meta/meta-error"
	metamongo "meta/meta-mongo"
	"meta/singleton"
)

type MigrateMarkDao struct {
	collection *mongo.Collection

	caches map[string]map[string]string // typeKey -> oldId -> newId
}

var singletonMigrateMarkDao = singleton.Singleton[MigrateMarkDao]{}

func GetMigrateMarkDao() *MigrateMarkDao {
	return singletonMigrateMarkDao.GetInstance(
		func() *MigrateMarkDao {
			mongoSubsystem := metamongo.GetSubsystem()
			if mongoSubsystem == nil {
				return nil
			}
			client := mongoSubsystem.GetClient()
			var MigrateMarkDao MigrateMarkDao
			MigrateMarkDao.collection = client.
				Database("didaoj").
				Collection("migrate_mark")
			return &MigrateMarkDao
		},
	)
}

func (d *MigrateMarkDao) GetMark(ctx context.Context, typeKey string, oldId string) (*string, error) {
	// Check cache first
	if d.caches != nil {
		if oldIdMap, exists := d.caches[typeKey]; exists {
			if newId, exists := oldIdMap[oldId]; exists {
				return &newId, nil // Return cached newId
			}
		}
	}
	filter := bson.M{
		"type":   typeKey,
		"old_id": oldId,
	}
	var result struct {
		NewId *string `bson:"new_id"`
	}
	err := d.collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil // No document found
		}
		return nil, metaerror.Wrap(err, "failed to find migrate mark")
	}
	if d.caches == nil {
		d.caches = make(map[string]map[string]string)
	}
	if _, exists := d.caches[typeKey]; !exists {
		d.caches[typeKey] = make(map[string]string)
	}
	if result.NewId != nil {
		d.caches[typeKey][oldId] = *result.NewId // Cache the result
	}
	return result.NewId, nil
}

func (d *MigrateMarkDao) Mark(ctx context.Context, typeKey string, oldId string, newId string) error {
	filter := bson.M{}
	filter["type"] = typeKey
	filter["old_id"] = oldId
	update := bson.M{
		"$set": bson.M{
			"type":   typeKey,
			"old_id": oldId,
			"new_id": newId,
		},
	}
	opts := options.Update().SetUpsert(true)
	_, err := d.collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return metaerror.Wrap(err, "failed to update problem crawl time")
	}
	return nil
}
