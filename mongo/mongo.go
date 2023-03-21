package mongo

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	// Path to the AWS CA file
	caFilePath = "../rds-combined-ca-cn-bundle.pem"

	// Timeout operations after N seconds
	connectTimeout  = 5
	queryTimeout    = 30
	username        = "docdb"
	password        = "prototyperoute"
	clusterEndpoint = "127.0.0.1:27017"

	tlsEnabled               = true
	tlsAllowInvalidHostnames = true
	directConnection         = true
	replicaSet               = "rs0"

	// Which instances to read from
	readPreference = "secondaryPreferred"

	connectionStringTemplate = "mongodb://%s:%s@%s/?tls=%v&replicaSet=%s&readpreference=%s&directConnection=%v"
)

type MongoDB struct {
	Client *mongo.Client
}

func NewMongoDB() (*MongoDB, error) {
	connectionURI := fmt.Sprintf(connectionStringTemplate, username, password,
		clusterEndpoint, tlsEnabled, replicaSet, readPreference, directConnection)

	tlsConfig, err := getCustomTLSConfig()
	if err != nil {
		return nil, fmt.Errorf("failed getting TLS configuration: %v", err)
	}

	client, err := mongo.NewClient(options.Client().ApplyURI(connectionURI).SetTLSConfig(tlsConfig))
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), connectTimeout*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to cluster: %v", err)
	}

	// Force a connection to verify our connection string
	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to ping cluster: %v", err)
	}

	fmt.Println("Connected to DocumentDB!")

	return &MongoDB{Client: client}, nil
}

func (m *MongoDB) InsertOne(collection *mongo.Collection, document interface{}) (interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout*time.Second)
	defer cancel()

	res, err := collection.InsertOne(ctx, document)
	if err != nil {
		return nil, fmt.Errorf("failed to insert document: %v", err)
	}

	return res.InsertedID, nil
}

func (m *MongoDB) InsertMany(collection *mongo.Collection, docs []interface{}) ([]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout*time.Second)
	defer cancel()

	res, err := collection.InsertMany(ctx, docs)
	if err != nil {
		return nil, fmt.Errorf("failed to insert documents: %v", err)
	}

	return res.InsertedIDs, nil
}

func (m *MongoDB) FindMany(collection *mongo.Collection,
	filter map[string]interface{}, projection map[string]interface{}) ([]bson.M, error) {

	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout*time.Second)
	defer cancel()

	findOptions := options.Find()

	if projection != nil {
		findOptions.SetProjection(projection)
	}

	cur, err := collection.Find(ctx, filter, findOptions)

	if err != nil {
		return nil, fmt.Errorf("failed to run find query: %v", err)
	}
	defer cur.Close(ctx)

	var results []bson.M

	for cur.Next(ctx) {
		var result bson.M
		err := cur.Decode(&result)
		log.Printf("Returned: %v", result)

		if err != nil {
			return nil, fmt.Errorf("failed to decode document: %v", err)
		}
		results = append(results, result)
	}

	if err := cur.Err(); err != nil {
		return nil, fmt.Errorf("failed to run find query: %v", err)
	}

	return results, nil
}

func (m *MongoDB) DeleteMany(collection *mongo.Collection, filter map[string]interface{}) error {

	// Delete the documents that match the filter criteria
	res, err := collection.DeleteMany(context.Background(), filter)
	if err != nil {
		return err
	}

	// Log the number of documents deleted
	log.Printf("Deleted %d documents from collection '%s'", res.DeletedCount, collection.Name())

	return nil
}

func getCustomTLSConfig() (*tls.Config, error) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: tlsAllowInvalidHostnames,
	}

	file, err := os.Open(caFilePath)
	if err != nil {
		return tlsConfig, err
	}
	defer file.Close()

	certs, err := io.ReadAll(file)
	if err != nil {
		return tlsConfig, err
	}

	tlsConfig.RootCAs = x509.NewCertPool()
	ok := tlsConfig.RootCAs.AppendCertsFromPEM(certs)

	if !ok {
		return tlsConfig, errors.New("failed parsing pem file")
	}

	return tlsConfig, nil
}
