package database

import (
"go.mongodb.org/mongo-driver/mongo"
"log"
"fmt"
"os"
"context"
"github.com/joho/godotenv"
"go.mongodb.org/mongo-driver/mongo/options"
"time"
)

func DbInstance() *mongo.Client{
	err:=godotenv.Load(".env")
	if err!=nil{
		log.Fatal("Error loading .env file")
	}
	mongoDb:=os.Getenv("MONGODB_URI")	

	if mongoDb == "" {
        log.Fatal("MONGODB_URI is not set in the environment")
	}

	ctx,cancel:=context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client,err:=mongo.Connect(ctx, options.Client().ApplyURI(mongoDb))

	if err!=nil{
		log.Fatal(err)
	}

	fmt.Println("Connected to MongoDB")
	return client
}

var Client *mongo.Client=DbInstance()

func OpenCollection(client *mongo.Client,collectionName string) *mongo.Collection{
	var collection *mongo.Collection=client.Database("cluster0").Collection(collectionName)
	return collection
}