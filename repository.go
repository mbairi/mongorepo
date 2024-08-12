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
	collection   *mongo.Collection
	idFieldIndex int
}

type QueryConfig struct {
	Query      bson.M
	Projection bson.M
	Sort       bson.M
	Pageable   []int
}

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

func (r *MongoRepository[T]) FindById(ctx context.Context, id primitive.ObjectID) (T, error) {
	var result T
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&result)
	return result, err
}

func (r *MongoRepository[T]) ExistsById(ctx context.Context, id primitive.ObjectID) (bool, error) {
	count, err := r.collection.CountDocuments(ctx, bson.M{"_id": id}, options.Count().SetLimit(1))
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *MongoRepository[T]) Save(ctx context.Context, item T) (T, error) {
	v := reflect.ValueOf(item)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	id := v.Field(r.idFieldIndex).Interface()
	result, err := r.collection.ReplaceOne(ctx, bson.M{"_id": id}, item, options.Replace().SetUpsert(true))
	if err != nil {
		return item, err
	}
	if result.UpsertedID != nil {
		v.Field(r.idFieldIndex).Set(reflect.ValueOf(result.UpsertedID))
	}
	return item, nil
}

func (r *MongoRepository[T]) SaveAll(ctx context.Context, items []T) ([]T, error) {
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
	result, err := r.collection.BulkWrite(ctx, writes)
	if err != nil {
		return items, err
	}
	for i, upsertedID := range result.UpsertedIDs {
		if upsertedID != nil {
			v := reflect.ValueOf(&items[i]).Elem()
			v.Field(r.idFieldIndex).Set(reflect.ValueOf(upsertedID))
		}
	}
	return items, nil
}

func (r *MongoRepository[T]) DeleteById(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (r *MongoRepository[T]) QueryOne(ctx context.Context, queryConfig QueryConfig) (T, error) {
	var result T
	findOptions := options.FindOne()
	if queryConfig.Projection != nil {
		findOptions.SetProjection(queryConfig.Projection)
	}
	err := r.collection.FindOne(ctx, queryConfig.Query, findOptions).Decode(&result)
	return result, err
}

func (r *MongoRepository[T]) QueryMany(ctx context.Context, queryConfig QueryConfig) ([]T, error) {
	var results []T
	findOptions := options.Find()
	if queryConfig.Sort != nil {
		findOptions.SetSort(queryConfig.Sort)
	}
	if queryConfig.Projection != nil {
		findOptions.SetProjection(queryConfig.Projection)
	}
	if len(queryConfig.Pageable) == 2 {
		findOptions.SetSkip(int64(queryConfig.Pageable[0]))
		findOptions.SetLimit(int64(queryConfig.Pageable[1]))
	}
	cursor, err := r.collection.Find(ctx, queryConfig.Query, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	err = cursor.All(ctx, &results)
	return results, err
}

func (r *MongoRepository[T]) AggregateOne(ctx context.Context, pipeline []bson.M) (bson.M, error) {
	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var result bson.M
	if cursor.Next(ctx) {
		err = cursor.Decode(&result)
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

func (r *MongoRepository[T]) AggregateMultiple(ctx context.Context, pipeline []bson.M) ([]bson.M, error) {
	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var results []bson.M
	err = cursor.All(ctx, &results)
	return results, err
}
