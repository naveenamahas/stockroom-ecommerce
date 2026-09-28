package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var users *mongo.Collection
var products *mongo.Collection
var orders *mongo.Collection
var rdb *redis.Client

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}
	mongoClient, err := mongo.Connect(options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal("MongoDB connection:", err)
	}
	if err = mongoClient.Ping(ctx, nil); err != nil {
		log.Fatal("MongoDB ping:", err)
	}
	databaseName := os.Getenv("MONGO_DB")
	if databaseName == "" {
		databaseName = "loginapp"
	}
	database := mongoClient.Database(databaseName)
	users = database.Collection("users")
	products = database.Collection("products")
	orders = database.Collection("orders")
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	rdb = redis.NewClient(&redis.Options{Addr: redisAddr})
	if err = rdb.Ping(ctx).Err(); err != nil {
		log.Fatal("Redis ping:", err)
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Println("Server: http://localhost:" + port)
	log.Fatal(http.ListenAndServe(":"+port, NewServer()))
}
