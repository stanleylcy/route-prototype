package main

import (
	"encoding/json"
	"log"
	"route-prototype/gopb"
	m "route-prototype/mongo"

	"go.mongodb.org/mongo-driver/bson"
)

func main() {
	// Set up MongoDB client and collection
	mongoDB, err := m.NewMongoDB()
	if err != nil {
		log.Fatal(err)
	}
	collection := mongoDB.Client.Database("sample-database").Collection("collection")

	// Insert sample-route document into the collection
	InsertedID, err := mongoDB.InsertOne(collection, sampleRoute())
	if err != nil {
		log.Fatalf("Failed to insert documents: %v", err)
	}
	log.Printf("InsertedID: %v", InsertedID)

	// Define filter to retrieve documents where gateway is 0.0.0.0
	filter := bson.M{"gateway": "0.0.0.0"}

	// Retrieve documents matching the filter
	results, err := mongoDB.FindMany(collection, filter, nil)
	if err != nil {
		log.Fatalf("Failed to retrieve documents: %v", err)
	}
	jsonRes, _ := json.Marshal(results)
	log.Printf("results: %v", string(jsonRes))

	// Delete all documents in the collection
	err = mongoDB.DeleteMany(collection, bson.M{})
	if err != nil {
		log.Fatalf("Failed to delete documents: %v", err)
	}
}

func sampleRoute() interface{} {
	route := gopb.Route{
		Destination: "192.168.79.0",
		Gateway:     "0.0.0.0",
		Genmask:     "255.255.255.0",
		Flags:       "U",
		Metric:      100,
		Ref:         0,
		Use:         0,
		Iface:       "ens33",
	}

	// Convert the Route struct to JSON
	jsonData, err := json.Marshal(&route)
	if err != nil {
		log.Fatalf("Error marshaling JSON data: %v", err)
	}

	// Convert the JSON data to BSON
	var bsonData interface{}
	err = bson.UnmarshalExtJSON(jsonData, true, &bsonData)
	if err != nil {
		log.Fatalf("Error unmarshaling BSON data: %v", err)
	}

	log.Printf("BSON data: %v", bsonData)

	return bsonData
}
