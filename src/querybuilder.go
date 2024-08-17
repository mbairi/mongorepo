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
	filter     string
	projection string
	sort       string
	context    context.Context
	pageable   [2]int
}

type Query struct {
	filter     bson.M
	projection bson.M
	sort       bson.D
	context    context.Context
	pageable   [2]int
}

func (q *QueryBuilder[T]) Filter(filter string) *QueryBuilder[T] {
	q.filter = filter
	return q
}

func (q *QueryBuilder[T]) Projection(projection string) *QueryBuilder[T] {
	q.projection = projection
	return q
}

func (q *QueryBuilder[T]) Sort(sort string) *QueryBuilder[T] {
	q.sort = sort
	return q
}

func (q *QueryBuilder[T]) Pageable(page [2]int) *QueryBuilder[T] {
	q.pageable = page
	return q
}

func (q *QueryBuilder[T]) Context(ctx context.Context) *QueryBuilder[T] {
	q.context = ctx
	return q
}

func (q *QueryBuilder[T]) Count(params ...interface{}) (int64, error) {
	query, err := q.ConstructQuery(params)
	if err != nil {
		return 0, err
	}
	return q.repo.Count(query)
}

func (q *QueryBuilder[T]) QueryOne(params ...interface{}) (T, error) {
	query, err := q.ConstructQuery(params...)
	if err != nil {
		var result T
		return result, err
	}

	return q.repo.QueryOne(query)
}

func (q *QueryBuilder[T]) QueryMany(params ...interface{}) ([]T, error) {
	query, err := q.ConstructQuery(params)
	if err != nil {
		var result []T
		return result, err
	}
	return q.repo.QueryMany(query)
}

func (q *QueryBuilder[T]) Delete(params ...interface{}) (int64, error) {
	query, err := q.ConstructQuery(params)
	if err != nil {
		return -1, err
	}
	return q.repo.Delete(query)
}

func (q *QueryBuilder[T]) ConstructQuery(params ...interface{}) (Query, error) {
	classic := Query{
		pageable: q.pageable,
		context:  q.context,
	}

	if q.filter != "" {
		queryStr := q.replaceParams(q.filter, params...)
		err := bson.UnmarshalExtJSON([]byte(queryStr), true, &classic.filter)
		if err != nil {
			return Query{}, fmt.Errorf("error parsing query: %w", err)
		}
	}

	if q.projection != "" {
		err := bson.UnmarshalExtJSON([]byte(q.projection), true, &classic.projection)
		if err != nil {
			return Query{}, fmt.Errorf("error parsing projection: %w", err)
		}
	}

	if q.sort != "" {
		var sortMap []map[string]int
		err := json.Unmarshal([]byte(q.sort), &sortMap)
		if err != nil {
			return Query{}, fmt.Errorf("error parsing sort: %w", err)
		}

		classic.sort = bson.D{}
		for _, m := range sortMap {
			for k, v := range m {
				classic.sort = append(classic.sort, bson.E{Key: k, Value: v})
			}
		}
	}

	return classic, nil
}

func (q *QueryBuilder[T]) replaceParams(query string, params ...interface{}) string {
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
