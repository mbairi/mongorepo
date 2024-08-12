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

// MongoRepository is a generic repository for MongoDB operations
type MongoRepository[T any] struct {
	collection   *mongo.Collection
	idFieldIndex int
}

type QueryConfig struct {
	Query      bson.M
	Projection bson.M
	Sort       bson.M
	Pageable   []int
}

// NewMongoRepository creates a new MongoRepository
func NewMongoRepository[T any](collection *mongo.Collection) (*MongoRepository[T], error) {
	repo := &MongoRepository[T]{
		collection: collection,
	}
	if err := repo.setIdField(); err != nil {
		return nil, err
	}
	if err := repo.ensureIndexes(); err != nil {
		return nil, err
	}
	return repo, nil
}

// setIdField finds the field with bson:"_id" tag
func (r *MongoRepository[T]) setIdField() error {
	var dummy T
	t := reflect.TypeOf(dummy)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	for i := 0; i < t.NumField(); i++ {
		if tag := t.Field(i).Tag.Get("bson"); tag == "_id" {
			r.idFieldIndex = i
			return nil
		}
	}
	return errors.New("type does not have a field with bson:\"_id\" tag")
}

// ensureIndexes creates indexes based on struct tags
func (r *MongoRepository[T]) ensureIndexes() error {
	var dummy T
	t := reflect.TypeOf(dummy)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	var indexes []mongo.IndexModel
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if tag := field.Tag.Get("index"); tag != "" {
			order := 1
			if strings.ToLower(tag) == "desc" {
				order = -1
			}
			index := mongo.IndexModel{
				Keys: bson.D{{field.Name, order}},
			}
			indexes = append(indexes, index)
		}
	}
	if len(indexes) > 0 {
		_, err := r.collection.Indexes().CreateMany(context.Background(), indexes)
		return err
	}
	return nil
}

// FindAll retrieves all documents in the collection
func (r *MongoRepository[T]) FindAll(ctx context.Context) ([]T, error) {
	var results []T
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	err = cursor.All(ctx, &results)
	return results, err
}

// FindById retrieves a document by its ID
func (r *MongoRepository[T]) FindById(ctx context.Context, id primitive.ObjectID) (T, error) {
	var result T
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&result)
	return result, err
}

// Save inserts or updates a document
func (r *MongoRepository[T]) Save(ctx context.Context, item T) error {
	v := reflect.ValueOf(item)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	id := v.Field(r.idFieldIndex).Interface()
	_, err := r.collection.ReplaceOne(ctx, bson.M{"_id": id}, item, options.Replace().SetUpsert(true))
	return err
}

// SaveAll inserts or updates multiple documents
func (r *MongoRepository[T]) SaveAll(ctx context.Context, items []T) error {
	var writes []mongo.WriteModel
	for _, item := range items {
		v := reflect.ValueOf(item)
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}
		id := v.Field(r.idFieldIndex).Interface()
		write := mongo.NewReplaceOneModel().SetFilter(bson.M{"_id": id}).SetReplacement(item).SetUpsert(true)
		writes = append(writes, write)
	}
	_, err := r.collection.BulkWrite(ctx, writes)
	return err
}

// DeleteById deletes a document by its ID
func (r *MongoRepository[T]) DeleteById(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

// FindOneCustom finds a single document based on a custom filter
func (r *MongoRepository[T]) FilterOne(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) (T, error) {
	var result T
	err := r.collection.FindOne(ctx, filter, opts...).Decode(&result)
	return result, err
}

// FindManyCustom finds multiple documents based on a custom filter
func (r *MongoRepository[T]) FilterMany(ctx context.Context, filter interface{}, opts ...*options.FindOptions) ([]T, error) {
	var results []T
	cursor, err := r.collection.Find(ctx, filter, opts...)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	err = cursor.All(ctx, &results)
	return results, err
}

func (r *MongoRepository[T]) DeleteOne(ctx context.Context, filter bson.M) error {
	_, err := r.collection.DeleteOne(ctx, filter)
	return err
}

// DeleteMany deletes multiple documents based on a filter
func (r *MongoRepository[T]) DeleteMany(ctx context.Context, filter bson.M) error {
	_, err := r.collection.DeleteMany(ctx, filter)
	return err
}

// Count returns the number of documents matching the filter
func (r *MongoRepository[T]) Count(ctx context.Context, filter bson.M) (int64, error) {
	return r.collection.CountDocuments(ctx, filter)
}

// Aggregate performs an aggregation pipeline
func (r *MongoRepository[T]) Aggregate(ctx context.Context, pipeline []bson.M) ([]bson.M, error) {
	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var results []bson.M
	err = cursor.All(ctx, &results)
	return results, err
}
