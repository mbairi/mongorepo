package example

import (
	"context"
	"fmt"
	mongorepo "mongorepo"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Person struct {
	ID    primitive.ObjectID `bson:"_id,omitempty"`
	Name  string             `bson:"name" index:"1"`
	Age   int                `bson:"age" index:"-1"`
	Email string             `bson:"email"`
}

type PersonRepository struct {
	*mongorepo.MongoRepository[Person]
}

// --- Constructing filter manually ---
func (r *PersonRepository) FindByNameAndAge(name string, age int) (Person, error) {
	queryConfig := mongorepo.ClassicQuery{
		Query: bson.M{"name": name, "age": age},
	}
	return r.QueryOne(context.TODO(), queryConfig)
}

func (r *PersonRepository) FindByAgeGreaterThan(age, page, limit int) ([]Person, error) {
	queryConfig := mongorepo.ClassicQuery{
		Query:      bson.M{"age": bson.M{"$gte": age}},
		Projection: bson.M{"name": 1},
		Sort:       bson.D{bson.E{Key: "age", Value: 1}},
		Pageable:   [2]int{0, 5},
	}
	return r.QueryMany(context.TODO(), queryConfig)
}

// --- Sending string filter like spring data for ease of use ---
func (r *PersonRepository) FindByAgeGreaterThanSimple(age, page, limit int) ([]Person, error) {
	simpleConfig := mongorepo.SimpleQuery{
		Query:    `{"age": {"$gte": ?1}}`,
		Sort:     `{"age":1}`,
		Pageable: [2]int{0, 5},
	}
	queryConfig, _ := simpleConfig.ToClassicQuery(age, page, limit)
	return r.QueryMany(context.TODO(), queryConfig)
}

func (r *PersonRepository) GroupByAge() {
	pipeline := []bson.M{
		{
			"$group": bson.M{
				"_id": "$age",
				"count": bson.M{
					"$sum": 1,
				},
			},
		},
		{
			"$sort": bson.M{
				"_id": 1,
			},
		},
	}
	results, err := r.AggregateMultiple(context.TODO(), pipeline)
	fmt.Println(results, err)
}

func (r *PersonRepository) AvgAge() {
	pipeline := []bson.M{
		{
			"$group": bson.M{
				"_id": nil,
				"avgAge": bson.M{
					"$avg": "$age",
				},
			},
		},
	}
	result, err := r.AggregateOne(context.TODO(), pipeline)
	fmt.Println(result, err)
}
