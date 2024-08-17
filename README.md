<p align="center">
<img width="330" height="110" src="logo.svg" border="0" alt="mongorepo"/> <br/>
<a><img src="https://img.shields.io/github/go-mod/go-version/mbairi/mongorepo" alt="License"></a> 
<a href="https://opensource.org/licenses/MIT"><img src="https://img.shields.io/badge/License-MIT-green.svg" alt="License"></a> 
</p>

<p align = "center">A Go library inspired by Spring's MongoRepository, reducing boilerplate and speeding up development by providing common MongoDB operations out of the box. </p>

# Usage

### Installation

```bash
go get github.com/mbairi/mongorepo
```

### Getting started

The repository for a document of your choice is created by embedding repo.MongoRepository that we provide using generics. When writing your constructor for the repository, instantiate it for that document type with the collection & embed it.

```go
import 	"github.com/mbairi/mongorepo/repo"

// Document struct we will be storing in db
type Person struct {
	ID    primitive.ObjectID `bson:"_id"`
	Name  string             `bson:"name"`
	Email string             `bson:"email"`
	Age   int64              `bson:"age"`
}

// Repository declaration
type PersonRepository struct {
	*repo.MongoRepository[Person]
}

func NewPersonRepository(collection *mongo.Collection) *PersonRepository {
	repoEmbedding, _ := repo.NewMongoRepository[Person](collection)
	return &PersonRepository{repoEmbedding}
}

func main() {
	// setup the databse & repository with manual injection
	client, _ := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://localhost:27017/testdb"))
	collection := client.Database("testdb").Collection("testcollection")
	personRepository := NewPersonRepository(collection)

	// save the person & fetch from database with updated ID just to test
	person := Person{Name: "Nitin", Email: "nitin.does.not.exist@gmail.com", Age: 26}
	savedPerson, _ := personRepository.Save(context.TODO(), person) // returns with updated ID ( input is NOT idempotent )
	foundPerson, _ := personRepository.FindById(context.TODO(), savedPerson.ID)

	fmt.Println(foundPerson)
}

```

### Default methods

Out of the box, these methods are provided by the library without any extra code.

| Function   | Description                                                     |
| ---------- | --------------------------------------------------------------- |
| Save       | Upserts a single item. If inserting, populates ID               |
| SaveAll    | Upserts all items in array. Populates ID for items if inserting |
| FindById   | Finds an item from collection matching \_id                     |
| FindByIds  | Finds items which match given list of ids                       |
| DeleteById | Deletes an object from collection matching \_id                 |
| FindAll    | Fetches all documents from given collection                     |
| ExistsById | Returns true if it finds an element with \_id                   |
| CountAll   | Returns count of all items present in collection                |

<br/>
The id related functions rely on the `bson:"\_id" tag in the struct defined for your document
<br/><br/>
Save & SaveAll are *NOT* idempotent, the items provided are updated with id if inserted & returns the same

### Simple Queries

```go
	foundItems, err := personRepository.QueryRunner().
		Filter(`{"age":{ "$gte": 30 }}`).
		Sort(`[{age:1}]`).
		QueryMany()
```

Chaining used to create the query

| Function   | Description                                                     |
| ---------- | --------------------------------------------------------------- |
| Filter     | basic filter for the operation                                  |
| Projection | sets the projection for the results                             |
| Sort       | accepts the sort order of items                                 |
| Pagination | accespts a [2]int{} with first number as page & second as limit |
| Context    | Sets context for query, uses default TODO() if not present      |

End functions to execute the query

| Function  | Description                                                     |
| --------- | --------------------------------------------------------------- |
| QueryOne  | basic filter for the operation                                  |
| QueryMany | sets the projection for the results                             |
| Count     | accepts the sort order of items                                 |
| Delete    | accespts a [2]int{} with first number as page & second as limit |

### Aggregates

Aggregation with multiple records as result:

```go
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
```

Aggregation with single record as result:

```go
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
```

### Simple Indexes

```go
type Person struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	Name        string             `bson:"name" index:"1, unique"`
	Age         int                `bson:"age" index:"-1"`
	Email 			string             `bson:"email" index:"text"`
}
```

## Extras

### Compound indexes

```go
type Person struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" cindex:"{age:1,name:-1};{age:1,email:1}"`
	Name        string             `bson:"name"`
	Age         int                `bson:"age"`
	Email 			string             `bson:"email" `
}
```
