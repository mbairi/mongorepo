package mongorepo

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
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	Name      string             `bson:"name"`
	Age       int                `bson:"age"`
	CreatedAt time.Time          `bson:"created_at"`
}

func setupTestRepo(t *testing.T) *MongoRepository[TestModel] {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://localhost:27017/testdb"))
	if err != nil {
		t.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	collection := client.Database("testdb").Collection("testcollection")

	// Drop the collection before each test to ensure a clean state
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
	ctx := context.TODO()

	newItem := TestModel{Name: "John Doe", Age: 30, CreatedAt: time.Now()}
	savedItem, err := repo.Save(ctx, newItem)
	if err != nil {
		t.Fatalf("Failed to save new item: %v", err)
	}
	if savedItem.ID.IsZero() {
		t.Fatalf("Expected non-zero ID for saved item")
	}

	savedItem.Name = "Jane Doe"
	updatedItem, err := repo.Save(ctx, savedItem)
	if err != nil {
		t.Fatalf("Failed to update item: %v", err)
	}
	if updatedItem.Name != "Jane Doe" {
		t.Fatalf("Expected updated name to be 'Jane Doe', got '%s'", updatedItem.Name)
	}
}

func TestFindById(t *testing.T) {
	repo := setupTestRepo(t)
	ctx := context.TODO()

	newItem := TestModel{Name: "Test User", Age: 25, CreatedAt: time.Now()}
	savedItem, err := repo.Save(ctx, newItem)
	if err != nil {
		t.Fatalf("Failed to save test item: %v", err)
	}

	foundItem, err := repo.FindById(ctx, savedItem.ID)
	if err != nil {
		t.Fatalf("Failed to find item by ID: %v", err)
	}
	if foundItem.ID != savedItem.ID {
		t.Fatalf("Expected found item ID to match saved item ID")
	}
}

func TestFindAll(t *testing.T) {
	repo := setupTestRepo(t)
	ctx := context.TODO()

	// Insert multiple test documents
	items := []TestModel{
		{Name: "User 1", Age: 25, CreatedAt: time.Now()},
		{Name: "User 2", Age: 30, CreatedAt: time.Now()},
		{Name: "User 3", Age: 35, CreatedAt: time.Now()},
	}
	for _, item := range items {
		_, err := repo.Save(ctx, item)
		if err != nil {
			t.Fatalf("Failed to save test item: %v", err)
		}
	}

	// Test finding all documents
	foundItems, err := repo.FindAll(ctx)
	if err != nil {
		t.Fatalf("Failed to find all items: %v", err)
	}
	if len(foundItems) != len(items) {
		t.Fatalf("Expected to find %d items, but found %d", len(items), len(foundItems))
	}
}

func TestSaveAll(t *testing.T) {
	repo := setupTestRepo(t)
	ctx := context.TODO()

	// Prepare multiple items to save
	items := []TestModel{
		{Name: "Bulk User 1", Age: 25, CreatedAt: time.Now()},
		{Name: "Bulk User 2", Age: 30, CreatedAt: time.Now()},
		{Name: "Bulk User 3", Age: 35, CreatedAt: time.Now()},
	}

	// Test saving multiple items
	savedItems, err := repo.SaveAll(ctx, items)
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
	ctx := context.TODO()

	// Insert a test document
	newItem := TestModel{Name: "Query Test", Age: 40, CreatedAt: time.Now()}
	_, err := repo.Save(ctx, newItem)
	if err != nil {
		t.Fatalf("Failed to save test item: %v", err)
	}

	// Test querying for one document
	query := ClassicQuery{
		Query: bson.M{"name": "Query Test"},
	}
	foundItem, err := repo.QueryOne(ctx, query)
	if err != nil {
		t.Fatalf("Failed to query one item: %v", err)
	}
	if foundItem.Name != "Query Test" || foundItem.Age != 40 {
		t.Fatalf("Query result does not match expected values")
	}
}

func TestQueryMany(t *testing.T) {
	repo := setupTestRepo(t)
	ctx := context.TODO()

	// Insert multiple test documents
	items := []TestModel{
		{Name: "Query Many 1", Age: 25, CreatedAt: time.Now()},
		{Name: "Query Many 2", Age: 30, CreatedAt: time.Now()},
		{Name: "Query Many 3", Age: 35, CreatedAt: time.Now()},
	}
	for _, item := range items {
		_, err := repo.Save(ctx, item)
		if err != nil {
			t.Fatalf("Failed to save test item: %v", err)
		}
	}

	// Test querying for multiple documents
	query := ClassicQuery{
		Query: bson.M{"age": bson.M{"$gte": 30}},
		Sort:  bson.D{bson.E{Key: "age", Value: 1}},
	}
	foundItems, err := repo.QueryMany(ctx, query)
	if err != nil {
		t.Fatalf("Failed to query many items: %v", err)
	}
	if len(foundItems) != 2 {
		t.Fatalf("Expected to find 2 items, but found %d", len(foundItems))
	}
	if foundItems[0].Age != 30 || foundItems[1].Age != 35 {
		t.Fatalf("Query results do not match expected values")
	}
}

func TestAggregateOne(t *testing.T) {
	repo := setupTestRepo(t)
	ctx := context.TODO()

	// Insert test documents
	items := []TestModel{
		{Name: "Agg One 1", Age: 25, CreatedAt: time.Now()},
		{Name: "Agg One 2", Age: 30, CreatedAt: time.Now()},
		{Name: "Agg One 3", Age: 35, CreatedAt: time.Now()},
	}
	for _, item := range items {
		_, err := repo.Save(ctx, item)
		if err != nil {
			t.Fatalf("Failed to save test item: %v", err)
		}
	}

	// Test aggregation for one result
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

	// Insert test documents
	items := []TestModel{
		{Name: "Agg Multi 1", Age: 25, CreatedAt: time.Now()},
		{Name: "Agg Multi 2", Age: 30, CreatedAt: time.Now()},
		{Name: "Agg Multi 3", Age: 35, CreatedAt: time.Now()},
	}
	for _, item := range items {
		_, err := repo.Save(ctx, item)
		if err != nil {
			t.Fatalf("Failed to save test item: %v", err)
		}
	}

	// Test aggregation for multiple results
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
	ctx := context.TODO()

	// Insert a test document
	newItem := TestModel{Name: "Delete Test", Age: 50, CreatedAt: time.Now()}
	savedItem, err := repo.Save(ctx, newItem)
	if err != nil {
		t.Fatalf("Failed to save test item: %v", err)
	}

	// Test deleting the document by ID
	err = repo.DeleteById(ctx, savedItem.ID)
	if err != nil {
		t.Fatalf("Failed to delete item by ID: %v", err)
	}

	// Verify the item is deleted
	_, err = repo.FindById(ctx, savedItem.ID)
	if err == nil {
		t.Fatalf("Expected item to be deleted, but it still exists")
	}
}
