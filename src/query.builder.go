package repo

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
)

type QueryBuilder[T any] struct {
	repo       *MongoRepository[T]
	filter     bson.M
	projection bson.M
	sort       bson.D
	context    context.Context
	pageable   [2]int
}

func (q *QueryBuilder[T]) Filter(filter string, params ...interface{}) *QueryBuilder[T] {
	queryStr := replaceParams(filter, params...)
	err := bson.UnmarshalExtJSON([]byte(queryStr), true, &q.filter)
	if err != nil {
		panic(err)
	}
	return q
}

func (q *QueryBuilder[T]) FilterB(filter bson.M) *QueryBuilder[T] {
	q.filter = filter
	return q
}

func (q *QueryBuilder[T]) Projection(projection string) *QueryBuilder[T] {
	err := bson.UnmarshalExtJSON([]byte(projection), true, &q.projection)
	if err != nil {
		panic(err)
	}
	return q
}

func (q *QueryBuilder[T]) ProjectionB(projection bson.M) *QueryBuilder[T] {
	q.projection = projection
	return q
}

func (q *QueryBuilder[T]) Sort(sort string) *QueryBuilder[T] {
	var sortMap []map[string]int
	err := json.Unmarshal([]byte(sort), &sortMap)
	if err != nil {
		panic(err)
	}

	q.sort = bson.D{}
	for _, m := range sortMap {
		for k, v := range m {
			q.sort = append(q.sort, bson.E{Key: k, Value: v})
		}
	}
	return q
}

func (q *QueryBuilder[T]) SortB(sort bson.D) *QueryBuilder[T] {
	q.sort = sort
	return q
}

func (q *QueryBuilder[T]) Pageable(pageable [2]int) *QueryBuilder[T] {
	q.pageable = pageable
	return q
}

func (q *QueryBuilder[T]) Context(ctx context.Context) *QueryBuilder[T] {
	q.context = ctx
	return q
}

func (q *QueryBuilder[T]) Count() (int64, error) {
	return q.repo.Count(q)
}

func (q *QueryBuilder[T]) QueryOne() (T, error) {
	return q.repo.QueryOne(q)
}

func (q *QueryBuilder[T]) QueryMany() ([]T, error) {
	return q.repo.QueryMany(q)
}

func (q *QueryBuilder[T]) Delete() (int64, error) {
	return q.repo.Delete(q)
}

func replaceParams(query string, params ...interface{}) string {
	for i, param := range params {
		placeholder := fmt.Sprintf("?%d", i+1)
		var replacement string

		// Handle strings separately to avoid wrapping them in arrays
		switch v := param.(type) {
		case string:
			replacement = fmt.Sprintf(`"%s"`, v)
		default:
			marshaledValue, err := json.Marshal(v)
			if err != nil {
				replacement = fmt.Sprintf("%v", v)
			} else {
				replacement = string(marshaledValue)
			}
		}

		query = strings.Replace(query, placeholder, replacement, -1)
	}
	return query
}
