# Installation

```bash
go get github.com/Manjunatha-b/mongorepo
```

# Usage

```go
type Person struct {
	ID    primitive.ObjectID `bson:"_id,omitempty"`
	Name  string             `bson:"name" index:"asc, unique"`
	Age   int                `bson:"age"`
	Email string             `bson:"email"`
}

type PersonRepository struct {
	*mongorepo.MongoRepository[Person]
}
```

### Inbuilt methods

| Function                                                  | Description                                                                           |
| --------------------------------------------------------- | ------------------------------------------------------------------------------------- |
| Save(ctx context.Context, item T )                        | If item is present in database, update it. Otherwise insert & return item with id set |
| SaveAll( ctx context.Context, item []T )                  | If item in array is new, set an id to it & insert to db, otherwise update.            |
| FindById( ctx context.Context, id primitive.ObjectId )    | Finds an item from collection matching \_id : given id                                |
| FindByIds( ctx context.Context, ids []primitve.ObjectId ) | Finds items whose ids match given ids                                                 |
| DeleteById( ctx context.Context, id primitive.ObjectId )  | Deletes an object from collection matching \_id : given id                            |
| FindAll( ctx context.Context )                            | Fetches all documents from given collection                                           |
| ExistsById ( ctx context.Context, id primitive.ObjectId ) | Returns true if it finds an element with \_id: given id                               |
| Count ( ctx context.Context, id primitive.ObjectId )      | Returns count of items present in collection                                          |

> These functions rely on the bson: "\_id" tag

> Save & SaveAll use reflection to setId on the bson:"\_id" tagged field if inserting

### Classic query

```go
func (r *PersonRepository) FindByAgeGreaterThan(age, page, limit int) ([]Person, error) {
	queryConfig := mongorepo.ClassicQuery{
		Query:      bson.M{"age": bson.M{"$gte": age}},
		Projection: bson.M{"name": 1},
		Sort:       bson.D{bson.E{Key: "name", Value: 1}},
		Pageable:   [2]int{0, 5},
	}
	return r.QueryMany(context.TODO(), queryConfig)
}

func (r *PersonRepository) FindByName(name string) ([]Person, error) {
	queryConfig := mongorepo.ClassicQuery{
		Query:      bson.M{"name": name},
	}
	return r.QueryMany(context.TODO(), queryConfig)
}
```

### Simple query

```go
func (r *PersonRepository) FindByAgeGreaterThan(age, page, limit int) ([]Person, error) {
	simpleConfig := mongorepo.ClassicQuery{
		Query:      `{ "age": { "$gte": ?1 }}`,
		Projection: `{ "name": 1 }`,
		Sort:       `[{ "name":1 }]`,
		Pageable:   [2]int{0, 5},
	}
  queryConfig,_ := simpleConfig.ToQueryConfig()
	return r.QueryMany(context.TODO(), queryConfig)
}
```

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
	Name        string             `bson:"name" index:"asc, unique"`
	Age         int                `bson:"age" index:"desc"`
	Description string             `bson:"description" index:"text"`
}
```

## Extras

### Compound indexes

```go
type Person struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" cindex:"{age:1,name:-1};{age:1,description:1}"`
	Name        string             `bson:"name" `
	Age         int                `bson:"age"`
	Description string             `bson:"description" `
}
```

### Fallback
