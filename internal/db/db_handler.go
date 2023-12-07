package db

import (
	"UNI/5sem/backend/9_pract/internal/config"
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net/http"
)

func ConnectToMongo() (*mongo.Client, error) {
	clientOptions := options.Client().ApplyURI(config.Load().MongoUrl)
	fmt.Println("MONGO_URL:", config.Load().MongoUrl)
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func DisconnectMongo(ctx context.Context, client *mongo.Client) {
	err := client.Disconnect(ctx)
	if err != nil {
		log.Printf("Disconnecting db error: %s", err)
	}
}

func NewBucket(data *mongo.Database) (bucket *gridfs.Bucket) {
	bucket, err := gridfs.NewBucket(data)
	if err != nil {
		log.Printf("Failed to create bucket:", http.StatusInternalServerError)
		return
	}
	return bucket
}
