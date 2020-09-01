package database

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"os"
)

var client *mongo.Client
var dbName string

type InsertResponse struct {
	ID  string `json:"_id,omitempty" bson:"_id,omitempty"`
	Err error	`json:"err,omitempty" bson:"err,omitempty"`
}

func InsertOne(collectionName string, object interface{}) InsertResponse {
	if client != nil {
		collection := client.Database(dbName).Collection(collectionName)

		result, err := collection.InsertOne(context.Background(), object)

		if err != nil {
			return InsertResponse{"", err}
		}

		return InsertResponse{result.InsertedID.(primitive.ObjectID).Hex(), err}
	}

	return InsertResponse{"", errors.New("MongoDB client not initialized yet")}
}

func FindOne(collectionName string, filter bson.D) (bson.M, error) {
	if client != nil {
		var result bson.M

		collection := client.Database(dbName).Collection(collectionName)

		if err := collection.FindOne(context.Background(), filter).Decode(&result); err != nil {
			if err == mongo.ErrNoDocuments {
				return nil, nil
			} else {
				return nil, err
			}
		}

		return result, nil

	}

	return nil, errors.New("MongoDB client not initialized yet")
}

func Disconnect() error {
	if client != nil {
		if err := client.Disconnect(context.TODO()); err != nil {
			return err
		}

		log.Println("Connection to MongoDB closed")
	}
	return nil
}

func Connect() error {
	mongoURI, uriExists := os.LookupEnv("MONGODB_URI")

	if !uriExists {
		return errors.New("MongoDB URI not declared; exiting")
	}

	var dbNameExists bool

	dbName, dbNameExists = os.LookupEnv("MONGODB_DATABASE_NAME")
	if !dbNameExists {
		return errors.New("MongoDB database name not declared; exiting")
	}

	clientOptions := options.Client().ApplyURI(mongoURI)

	var connErr error = nil

	client, connErr = mongo.Connect(context.TODO(), clientOptions)

	if connErr != nil {
		return connErr
	}

	if pingErr := client.Ping(context.TODO(), nil); pingErr != nil {
		return pingErr
	}

	log.Println("Connected to MongoDB!")

	return nil
}
