package repo

import (
	"context"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type TestModel struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" cindex:"{name:1,age:1};{age:1,created_at:1}"`
	Name      string             `bson:"name" index:"1"`
	Age       int                `bson:"age"`
	CreatedAt time.Time          `bson:"created_at"`
}

func setupTestRepo(t *testing.T) *MongoRepository[TestModel] {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://localhost:27017/testdb"))
	if err != nil {
		t.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	collection := client.Database("testdb").Collection("testcollection")

	err = collection.Drop(context.TODO())
	if err != nil {
		t.Fatalf("Failed to drop collection: %v", err)
	}

	repo, err := NewMongoRepository[TestModel](collection)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}
	return repo
}

func TestSave(t *testing.T) {
	repo := setupTestRepo(t)

	newItem := TestModel{Name: "John Doe", Age: 30, CreatedAt: time.Now()}
	savedItem, err := repo.Save(newItem)
	if err != nil {
		t.Fatalf("Failed to save new item: %v", err)
	}
	if savedItem.ID.IsZero() {
		t.Fatalf("Expected non-zero ID for saved item")
	}

	savedItem.Name = "Jane Doe"
	updatedItem, err := repo.Save(savedItem)
	if err != nil {
		t.Fatalf("Failed to update item: %v", err)
	}
	if updatedItem.Name != "Jane Doe" {
		t.Fatalf("Expected updated name to be 'Jane Doe', got '%s'", updatedItem.Name)
	}

	count, err := repo.CountAll()
	if err != nil {
		t.Fatalf("Failed to get count of items after saving: %v", err)
	}
	if count != 1 {
		t.Fatalf("Expected db count to be 1 got %d", count)
	}
}

func TestFindById(t *testing.T) {
	repo := setupTestRepo(t)

	newItem := TestModel{Name: "Test User", Age: 25, CreatedAt: time.Now()}
	savedItem, err := repo.Save(newItem)
	if err != nil {
		t.Fatalf("Failed to save test item: %v", err)
	}

	foundItem, err := repo.FindById(savedItem.ID)
	if err != nil {
		t.Fatalf("Failed to find item by ID: %v", err)
	}
	if foundItem.ID != savedItem.ID {
		t.Fatalf("Expected found item ID to match saved item ID")
	}
}

func TestFindByIds(t *testing.T) {
	repo := setupTestRepo(t)
	items := []TestModel{
		{Name: "User 1", Age: 25, CreatedAt: time.Now()},
		{Name: "User 2", Age: 30, CreatedAt: time.Now()},
		{Name: "User 3", Age: 35, CreatedAt: time.Now()},
	}

	savedItems, err := repo.SaveAll(items)
	if err != nil {
		t.Fatalf("Faild to save items: %v", err)
	}

	ids := []primitive.ObjectID{savedItems[0].ID, savedItems[1].ID}

	foundItems, err := repo.FindByIds(ids)
	if err != nil {
		t.Fatalf("Failed to find all items: %v", err)
	}
	if len(foundItems) != 2 {
		t.Fatalf("Expected to find 2 items, but found %d", len(foundItems))
	}
}

func CountAll(t *testing.T) {
	repo := setupTestRepo(t)
	items := []TestModel{
		{Name: "User 1", Age: 25, CreatedAt: time.Now()},
		{Name: "User 2", Age: 30, CreatedAt: time.Now()},
		{Name: "User 3", Age: 35, CreatedAt: time.Now()},
	}

	_, err := repo.SaveAll(items)
	if err != nil {
		t.Fatalf("Faild to save items: %v", err)
	}

	count, err := repo.CountAll()
	if err != nil {
		t.Fatalf("Failed to find all items: %v", err)
	}
	if count != 3 {
		t.Fatalf("Expected to count 3 items, but found %d", count)
	}
}

func TestFindAll(t *testing.T) {
	repo := setupTestRepo(t)
	items := []TestModel{
		{Name: "User 1", Age: 25, CreatedAt: time.Now()},
		{Name: "User 2", Age: 30, CreatedAt: time.Now()},
		{Name: "User 3", Age: 35, CreatedAt: time.Now()},
	}

	_, err := repo.SaveAll(items)
	if err != nil {
		t.Fatalf("Faield to save items: %v", err)
	}

	foundItems, err := repo.FindAll()
	if err != nil {
		t.Fatalf("Failed to find all items: %v", err)
	}
	if len(foundItems) != len(items) {
		t.Fatalf("Expected to find %d items, but found %d", len(items), len(foundItems))
	}
}

func TestExistsById(t *testing.T) {
	repo := setupTestRepo(t)

	newItem := TestModel{Name: "John Doe", Age: 30, CreatedAt: time.Now()}
	savedItem, err := repo.Save(newItem)
	if err != nil {
		t.Fatalf("Failed to save new item: %v", err)
	}
	if savedItem.ID.IsZero() {
		t.Fatalf("Expected non-zero ID for saved item")
	}

	exists, err := repo.ExistsById(savedItem.ID)

	if err != nil {
		t.Fatalf("Failed to check if item exists: %v", err)
	}

	if !exists {
		t.Fatalf("Expected to find existing element")
	}
}

func TestSaveAll(t *testing.T) {
	repo := setupTestRepo(t)
	items := []TestModel{
		{Name: "Bulk User 1", Age: 25, CreatedAt: time.Now()},
		{Name: "Bulk User 2", Age: 30, CreatedAt: time.Now()},
		{Name: "Bulk User 3", Age: 35, CreatedAt: time.Now()},
	}

	// Test saving multiple items
	savedItems, err := repo.SaveAll(items)
	if err != nil {
		t.Fatalf("Failed to save multiple items: %v", err)
	}
	if len(savedItems) != len(items) {
		t.Fatalf("Expected to save %d items, but saved %d", len(items), len(savedItems))
	}
	for _, item := range savedItems {
		if item.ID.IsZero() {
			t.Fatalf("Expected non-zero ID for saved item")
		}
	}
}

func TestQueryOne(t *testing.T) {
	repo := setupTestRepo(t)
	newItem := TestModel{Name: "Query Test", Age: 40, CreatedAt: time.Now()}
	_, err := repo.Save(newItem)
	if err != nil {
		t.Fatalf("Failed to save test item: %v", err)
	}

	foundItem, err := repo.QueryRunner().
		Filter(`{"name":?1}`, newItem.Name).
		QueryOne()

	if err != nil {
		t.Fatalf("Failed to query one item: %v", err)
	}
	if foundItem.Name != "Query Test" || foundItem.Age != 40 {
		t.Fatalf("Query result does not match expected values")
	}
}

func TestQueryMany(t *testing.T) {
	repo := setupTestRepo(t)

	items := []TestModel{
		{Name: "Query Many 1", Age: 25, CreatedAt: time.Now()},
		{Name: "Query Many 2", Age: 30, CreatedAt: time.Now()},
		{Name: "Query Many 3", Age: 35, CreatedAt: time.Now()},
	}

	_, err := repo.SaveAll(items)
	if err != nil {
		t.Fatalf("Failed to save multiple items: %v", err)
	}

	filterAge := 30
	foundItems, err := repo.QueryRunner().
		Filter(`{"age":{ "$gte": ?1 }}`, filterAge).
		Sort(`[{"age":1}]`).
		Projection(`{"name":1, "age":1}`).
		QueryMany()

	if err != nil {
		t.Fatalf("Failed to query many items: %v", err)
	}
	if len(foundItems) != 2 {
		t.Fatalf("Expected to find 2 items, but found %d", len(foundItems))
	}
	if foundItems[0].Age != 30 || foundItems[1].Age != 35 {
		t.Fatalf("Query results do not match expected values")
	}
	if !foundItems[0].CreatedAt.IsZero() {
		t.Fatalf("Expected createdAt to be zero from projection, but found %s", foundItems[0].CreatedAt)
	}
}

func TestCount(t *testing.T) {
	repo := setupTestRepo(t)

	items := []TestModel{
		{Name: "Query Many 1", Age: 25, CreatedAt: time.Now()},
		{Name: "Query Many 2", Age: 30, CreatedAt: time.Now()},
		{Name: "Query Many 3", Age: 35, CreatedAt: time.Now()},
	}

	_, err := repo.SaveAll(items)
	if err != nil {
		t.Fatalf("Failed to save multiple items: %v", err)
	}

	count, err := repo.QueryRunner().
		Filter(`{"age":{ "$gte": 30 }}`).
		Count()

	if err != nil {
		t.Fatalf("Failed to query many items: %v", err)
	}
	if count != 2 {
		t.Fatalf("Expected to find 2 items, but found %d", count)
	}
}

func TestDelete(t *testing.T) {
	repo := setupTestRepo(t)

	items := []TestModel{
		{Name: "Query Many 1", Age: 25, CreatedAt: time.Now()},
		{Name: "Query Many 2", Age: 30, CreatedAt: time.Now()},
		{Name: "Query Many 3", Age: 35, CreatedAt: time.Now()},
	}

	_, err := repo.SaveAll(items)
	if err != nil {
		t.Fatalf("Failed to save multiple items: %v", err)
	}

	count, err := repo.QueryRunner().
		Filter(`{"age":{ "$gte": 30 }}`).
		Delete()

	if err != nil {
		t.Fatalf("Failed to query many items: %v", err)
	}
	if count != 2 {
		t.Fatalf("Expected to find 2 items, but found %d", count)
	}

	remCount, err := repo.CountAll()
	if remCount != 1 {
		t.Fatalf("Expected to find 1 item remaining in repo, but found %d", remCount)
	}
}

func TestAggregateOne(t *testing.T) {
	repo := setupTestRepo(t)
	ctx := context.TODO()

	items := []TestModel{
		{Name: "Agg One 1", Age: 25, CreatedAt: time.Now()},
		{Name: "Agg One 2", Age: 30, CreatedAt: time.Now()},
		{Name: "Agg One 3", Age: 35, CreatedAt: time.Now()},
	}
	for _, item := range items {
		_, err := repo.Save(item)
		if err != nil {
			t.Fatalf("Failed to save test item: %v", err)
		}
	}

	pipeline := []bson.M{
		{"$group": bson.M{"_id": nil, "avgAge": bson.M{"$avg": "$age"}}},
	}
	result, err := repo.AggregateOne(ctx, pipeline)
	if err != nil {
		t.Fatalf("Failed to aggregate one: %v", err)
	}
	avgAge := result["avgAge"].(float64)
	if avgAge != 30 {
		t.Fatalf("Expected average age to be 30, got %f", avgAge)
	}
}

func TestAggregateMultiple(t *testing.T) {
	repo := setupTestRepo(t)
	ctx := context.TODO()
	items := []TestModel{
		{Name: "Agg Multi 1", Age: 25, CreatedAt: time.Now()},
		{Name: "Agg Multi 2", Age: 30, CreatedAt: time.Now()},
		{Name: "Agg Multi 3", Age: 35, CreatedAt: time.Now()},
	}
	for _, item := range items {
		_, err := repo.Save(item)
		if err != nil {
			t.Fatalf("Failed to save test item: %v", err)
		}
	}

	pipeline := []bson.M{
		{"$group": bson.M{"_id": "$age", "count": bson.M{"$sum": 1}}},
		{"$sort": bson.M{"_id": 1}},
	}
	results, err := repo.AggregateMultiple(ctx, pipeline)
	if err != nil {
		t.Fatalf("Failed to aggregate multiple: %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("Expected 3 results, got %d", len(results))
	}
	for _, result := range results {
		if result["count"].(int32) != 1 {
			t.Fatalf("Expected count to be 1 for each age group")
		}
	}
}

func TestDeleteById(t *testing.T) {
	repo := setupTestRepo(t)

	newItem := TestModel{Name: "Delete Test", Age: 50, CreatedAt: time.Now()}
	savedItem, err := repo.Save(newItem)
	if err != nil {
		t.Fatalf("Failed to save test item: %v", err)
	}

	err = repo.DeleteById(savedItem.ID)
	if err != nil {
		t.Fatalf("Failed to delete item by ID: %v", err)
	}

	_, err = repo.FindById(savedItem.ID)
	if err == nil {
		t.Fatalf("Expected item to be deleted, but it still exists")
	}
}
