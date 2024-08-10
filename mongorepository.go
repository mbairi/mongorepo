package mongorepo

import (
	"context"
	"errors"
	"reflect"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoRepository[T any] struct {
	*mongo.Collection
	idFieldIndex int
}

func NewMongoRepository[T any](collection *mongo.Collection) (*MongoRepository[T], error) {
	repo := &MongoRepository[T]{
		Collection:   collection,
		idFieldIndex: -1,
	}
	if err := repo.setIdField(); err != nil {
		return nil, err
	}
	if err := repo.ensureIndexes(); err != nil {
		return nil, err
	}
	return repo, nil
}

func (r *MongoRepository[T]) setIdField() error {
	var dummy T
	t := reflect.TypeOf(dummy)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if tag := field.Tag.Get("bson"); tag != "" {
			parts := strings.Split(tag, ",")
			for _, part := range parts {
				if strings.TrimSpace(part) == "_id" {
					r.idFieldIndex = i
					return nil
				}
			}
		}
	}

	return errors.New("type does not have a field with bson:\"_id\" tag")
}

func (r *MongoRepository[T]) ensureIndexes() error {
	var dummy T
	t := reflect.TypeOf(dummy)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	var indexes []mongo.IndexModel

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if tag := field.Tag.Get("repo"); tag != "" {
			parts := strings.Split(tag, ",")
			for _, part := range parts {
				if strings.HasPrefix(part, "index") {
					indexParts := strings.Split(part, "_")
					if len(indexParts) != 2 {
						return errors.New("invalid index tag format")
					}
					order := 1
					if indexParts[1] == "desc" {
						order = -1
					}
					index := mongo.IndexModel{
						Keys: bson.D{{field.Name, order}},
					}
					indexes = append(indexes, index)
				}
			}
		}
	}

	if len(indexes) > 0 {
		_, err := r.Collection.Indexes().CreateMany(context.Background(), indexes)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *MongoRepository[T]) Save(item *T) (*T, error) {
	v := reflect.ValueOf(item).Elem()
	idField := v.Field(r.idFieldIndex)

	if !idField.IsValid() {
		return nil, errors.New("item does not have a valid _id field")
	}

	filter := bson.M{"_id": idField.Interface()}
	opts := options.Replace().SetUpsert(true)
	result, err := r.Collection.ReplaceOne(context.Background(), filter, item, opts)
	if err != nil {
		return nil, err
	}

	if result.UpsertedID != nil {
		idField.Set(reflect.ValueOf(result.UpsertedID))
	}

	return item, nil
}

func (r *MongoRepository[T]) FindById(id primitive.ObjectID) (T, error) {
	var data T
	filter := bson.M{"_id": id}
	res := r.Collection.FindOne(context.TODO(), filter)
	res.Decode(&data)
	return data, nil
}
