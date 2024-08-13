package example

import (
	"context"
	mongorepo "mongorepo/main"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Person struct {
	ID    primitive.ObjectID `bson:"_id,omitempty"`
	Name  string             `bson:"name" index:"asc, unique"`
	Age   int                `bson:"age"`
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
		Query:    bson.M{"age": bson.M{"$gte": age}},
		Sort:     bson.D{bson.E{Key: "age", Value: 1}},
		Pageable: [2]int{0, 5},
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
