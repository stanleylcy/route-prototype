# route-prototype
## Connecting to an Amazon DocumentDB Cluster from Outside an Amazon VPC

You might typically connect to an EC2 instance using the following command.
``` sh
ssh -i "ec2Access.pem" ubuntu@ec2-34-229-221-164.compute-1.amazonaws.com
```
If so, you can set up an SSH tunnel to the Amazon DocumentDB cluster ```sample-cluster.node.us-east-1.docdb.amazonaws.com``` by running the following command on your local computer. The ```-L``` flag is used for forwarding a local port. When using an SSH tunnel, we recommend that you connect to your cluster using the cluster endpoint and do not attempt to connect in replica set mode (i.e., specifying ```replicaSet=rs0``` in your connection string) as it will result in an error.
``` sh
ssh -i "ec2Access.pem" -L 27017:sample-cluster.node.us-east-1.docdb.amazonaws.com:27017 ubuntu@ec2-34-229-221-164.compute-1.amazonaws.com -N 
```
After the SSH tunnel is created, any commands that you issue to ```localhost:27017``` are forwarded to the Amazon DocumentDB cluster ```sample-cluster``` running in the Amazon VPC. If Transport Layer Security (TLS) is enabled on your Amazon DocumentDB cluster, you need to download the public key for Amazon DocumentDB from . The following operation downloads this file:
``` sh
wget https://s3.cn-north-1.amazonaws.com.cn/rds-downloads/rds-combined-ca-cn-bundle.pem
```
To connect to your Amazon DocumentDB cluster from outside the Amazon VPC, use the following command.
``` sh
mongosh --sslAllowInvalidHostnames --ssl --host localhost:27017 --sslCAFile rds-combined-ca-cn-bundle.pem --username <yourUsername> --password <yourPassword> 
```
## Sample Code
``` go
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
```
