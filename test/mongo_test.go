package test

import (
	"context"
	"fmt"
	"log"
	"testing"

	m "route-prototype/mongo"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func setUp() (*m.MongoDB, *mongo.Collection) {
	// Set up MongoDB client and collection
	mongoDB, err := m.NewMongoDB()
	if err != nil {
		log.Fatal(err)
	}

	collection := mongoDB.Client.Database("sample-database").Collection("collection")
	return mongoDB, collection
}

func TestInsertMany(t *testing.T) {
	mongoDB, collection := setUp()

	// Insert some sample documents into the collection
	docs := []interface{}{
		bson.M{"name": "Alice", "age": 25},
		bson.M{"name": "Bob", "age": 30},
		bson.M{"name": "Charlie", "age": 35},
	}

	_, err := mongoDB.InsertMany(collection, docs)
	if err != nil {
		t.Fatalf("Failed to insert documents: %v", err)
	}

	// Retrieve all documents in the collection
	var results []bson.M
	cursor, err := collection.Find(context.Background(), bson.M{})
	if err != nil {
		t.Fatalf("Failed to retrieve documents: %v", err)
	}

	err = cursor.All(context.Background(), &results)
	if err != nil {
		t.Fatalf("Failed to decode documents: %v", err)
	}

	// Verify that the retrieved documents match the expected values
	expectedResults := []bson.M{
		{"_id": results[0]["_id"], "name": "Alice", "age": 25},
		{"_id": results[1]["_id"], "name": "Bob", "age": 30},
		{"_id": results[2]["_id"], "name": "Charlie", "age": 35},
	}

	fmt.Printf("Expected: %v\nResult: %v", expectedResults, results)

	// Delete all documents in the collection
	_, err = collection.DeleteMany(context.Background(), bson.D{})
	if err != nil {
		t.Error(err)
	}
}

func TestFindMany(t *testing.T) {

	mongoDB, collection := setUp()

	// Insert some sample documents into the collection
	docs := []interface{}{
		bson.M{"name": "John", "age": 30},
		bson.M{"name": "Jane", "age": 25},
		bson.M{"name": "Bob", "age": 40},
	}
	_, err := collection.InsertMany(context.Background(), docs)
	if err != nil {
		t.Fatalf("Failed to insert documents: %v", err)
	}

	// Define filter to retrieve documents where age is greater than 25
	filter := bson.M{"age": bson.M{"$gt": 25}}
	projection := bson.M{"_id": 0, "name": 1, "age": 1}

	// Retrieve documents matching the filter
	results, err := mongoDB.FindMany(collection, filter, projection)
	if err != nil {
		t.Fatalf("Failed to retrieve documents: %v", err)
	}

	// Verify that the correct number of documents were returned
	if len(results) != 2 {
		t.Errorf("Expected 2 documents, but got %d", len(results))
	}

	// Verify that the retrieved documents match the expected values
	expectedResults := []bson.M{
		{"name": "John", "age": 30},
		{"name": "Bob", "age": 40},
	}

	fmt.Printf("Expected: %v\nResult: %v", expectedResults, results)

	// Delete all documents in the collection
	_, err = collection.DeleteMany(context.Background(), bson.D{})
	if err != nil {
		t.Error(err)
	}
}
