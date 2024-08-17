package repo

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
)

type CQueryBuilder[T any] struct {
	repo       *MongoRepository[T]
	filter     bson.M
	projection bson.M
	sort       bson.D
	context    context.Context
	pageable   [2]int
}

func (q *CQueryBuilder[T]) Filter(filter bson.M) *CQueryBuilder[T] {
	q.filter = filter
	return q
}

func (q *CQueryBuilder[T]) Projection(projection bson.M) *CQueryBuilder[T] {
	q.projection = projection
	return q
}

func (q *CQueryBuilder[T]) Sort(sort bson.D) *CQueryBuilder[T] {
	q.sort = sort
	return q
}

func (q *CQueryBuilder[T]) Pageable(page [2]int) *CQueryBuilder[T] {
	q.pageable = page
	return q
}

func (q *CQueryBuilder[T]) Context(ctx context.Context) *CQueryBuilder[T] {
	q.context = ctx
	return q
}

func (q *CQueryBuilder[T]) Count(params ...interface{}) (int64, error) {
	query := q.ToQuery()
	return q.repo.Count(query)
}

func (q *CQueryBuilder[T]) QueryOne(params ...interface{}) (T, error) {
	query := q.ToQuery()

	return q.repo.QueryOne(query)
}

func (q *CQueryBuilder[T]) QueryMany(params ...interface{}) ([]T, error) {
	query := q.ToQuery()
	return q.repo.QueryMany(query)
}

func (q *CQueryBuilder[T]) Delete(params ...interface{}) (int64, error) {
	query := q.ToQuery()
	return q.repo.Delete(query)
}

func (cq *CQueryBuilder[T]) ToQuery() Query {
	return Query{
		filter:     cq.filter,
		projection: cq.projection,
		sort:       cq.sort,
		context:    cq.context,
		pageable:   cq.pageable,
	}
}
