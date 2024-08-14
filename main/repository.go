package mongorepo

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strconv"
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

func NewMongoRepository[T any](
	collection *mongo.Collection,
) (*MongoRepository[T], error) {
	repo := &MongoRepository[T]{
		collection: collection,
	}
	if err := repo.setIdField(); err != nil {
		return nil, err
	}
	if err := repo.ensureSimpleIndexes(); err != nil {
		return nil, err
	}
	if err := repo.ensureCompoundIndex(); err != nil {
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
		if tag := t.Field(i).Tag.Get("bson"); tag != "" {
			tags := strings.Split(tag, ",")
			for _, t := range tags {
				if strings.TrimSpace(t) == "_id" {
					r.idFieldIndex = i
					return nil
				}
			}
		}
	}

	return errors.New("type does not have a field with bson:\"_id\" tag")
}

func (r *MongoRepository[T]) ensureSimpleIndexes() error {
	var dummy T
	t := reflect.TypeOf(dummy)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	var indexes []mongo.IndexModel
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldName := field.Name
		if tag := field.Tag.Get("bson"); tag != "" {
			splitTags := strings.Split(tag, ",")
			fieldName = splitTags[0]
		}
		if tag := field.Tag.Get("index"); tag != "" {
			splitTags := strings.Split(tag, ",")
			var indexType interface{}
			isUnique := false
			for _, splitTag := range splitTags {
				if splitTag == "unique" {
					isUnique = true
				} else if splitTag == "1" || splitTag == "-1" {
					indexType, _ = strconv.Atoi(splitTag)
				} else {
					indexType = splitTag
				}
			}
			indexOptions := options.IndexOptions{Unique: &isUnique}
			index := mongo.IndexModel{
				Keys:    bson.D{{Key: fieldName, Value: indexType}},
				Options: &indexOptions,
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

func (r *MongoRepository[T]) ensureCompoundIndex() error {
	var t T
	elemType := reflect.TypeOf(t)
	field := elemType.Field(r.idFieldIndex)
	cindexTag := field.Tag.Get("cindex")
	if cindexTag == "" {
		return nil // No index to create
	}

	cleanedCindex := strings.ReplaceAll(cindexTag, "{", "")
	cleanedCindex = strings.ReplaceAll(cleanedCindex, "}", "")
	indexes := strings.Split(cleanedCindex, ";")

	for _, index := range indexes {
		indexKeys := bson.D{}
		parts := strings.Split(index, ",")

		for _, part := range parts {
			kv := strings.Split(part, ":")
			if len(kv) != 2 {
				return fmt.Errorf("invalid compound index format: %s", part)
			}

			fieldName := kv[0]
			order, err := strconv.Atoi(kv[1])
			if err != nil {
				return fmt.Errorf("invalid compound index order: %s", kv[1])
			}

			indexKeys = append(indexKeys, bson.E{Key: fieldName, Value: order})
		}

		indexModel := mongo.IndexModel{
			Keys: indexKeys,
		}

		_, err := r.collection.Indexes().CreateOne(context.TODO(), indexModel)
		if err != nil {
			return fmt.Errorf("failed to create index: %v", err)
		}
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

func (r *MongoRepository[T]) FindByIds(ctx context.Context, ids []primitive.ObjectID) ([]T, error) {
	var results []T
	cursor, err := r.collection.Find(ctx, bson.M{"_id": bson.M{"$in": ids}})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	err = cursor.All(ctx, &results)
	return results, err
}

func (r *MongoRepository[T]) ExistsById(ctx context.Context, id primitive.ObjectID) (bool, error) {
	count, err := r.collection.CountDocuments(ctx, bson.M{"_id": id}, options.Count().SetLimit(1))
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *MongoRepository[T]) Count(ctx context.Context) (int64, error) {
	count, err := r.collection.CountDocuments(ctx, bson.M{}, options.Count().SetLimit(1))
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *MongoRepository[T]) Save(ctx context.Context, item T) (T, error) {
	v := reflect.ValueOf(&item).Elem()
	idField := v.Field(r.idFieldIndex)
	id := idField.Interface().(primitive.ObjectID)
	if id.IsZero() {
		result, err := r.collection.InsertOne(ctx, item)
		if err != nil {
			return item, err
		}
		idField.Set(reflect.ValueOf(result.InsertedID.(primitive.ObjectID)))
	} else {
		_, err := r.collection.ReplaceOne(ctx, bson.M{"_id": id}, item, options.Replace().SetUpsert(true))
		if err != nil {
			return item, err
		}
	}

	return item, nil
}

func (r *MongoRepository[T]) SaveAll(ctx context.Context, items []T) ([]T, error) {
	var writes []mongo.WriteModel
	for i := range items {
		v := reflect.ValueOf(&items[i]).Elem()
		idField := v.Field(r.idFieldIndex)
		id := idField.Interface().(primitive.ObjectID)
		if id.IsZero() {
			id = primitive.NewObjectID()
			idField.Set(reflect.ValueOf(id))
		}

		write := mongo.NewReplaceOneModel().
			SetFilter(bson.M{"_id": id}).
			SetReplacement(items[i]).
			SetUpsert(true)
		writes = append(writes, write)
	}

	_, err := r.collection.BulkWrite(ctx, writes)
	if err != nil {
		return items, err
	}
	return items, nil
}

func (r *MongoRepository[T]) DeleteById(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (r *MongoRepository[T]) QueryOne(ctx context.Context, queryConfig ClassicQuery) (T, error) {
	var result T
	findOptions := options.FindOne()
	if queryConfig.Projection != nil {
		findOptions.SetProjection(queryConfig.Projection)
	}
	err := r.collection.FindOne(ctx, queryConfig.Query, findOptions).Decode(&result)
	return result, err
}

func (r *MongoRepository[T]) QueryMany(ctx context.Context, queryConfig ClassicQuery) ([]T, error) {
	var results []T
	findOptions := options.Find()
	if queryConfig.Sort != nil {
		findOptions.SetSort(queryConfig.Sort)
	}
	if queryConfig.Projection != nil {
		findOptions.SetProjection(queryConfig.Projection)
	}
	if len(queryConfig.Pageable) == 2 {
		findOptions.SetSkip(int64(queryConfig.Pageable[1] * queryConfig.Pageable[0]))
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
