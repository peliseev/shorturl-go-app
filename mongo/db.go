package mongo

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func Open(url string) *mongo.Client {
	ctx := context.Background()
	clientOptions := options.Client().ApplyURI(url)

	mongoClient, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	err = mongoClient.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	return mongoClient
}
