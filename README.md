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

### Simple query

### Aggregates

### Indexes

### Fallback
