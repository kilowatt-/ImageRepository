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

type UpdateResponse struct {
	Modified int64
	Matched int64
}

func DeleteMany(collectionName string, filter bson.D) (int64, error) {
	if client != nil {
		collection := client.Database(dbName).Collection(collectionName)

		result, err := collection.DeleteMany(context.Background(), filter)

		if err != nil {
			return -1, err
		}

		return result.DeletedCount, nil
	}

	return -1, errors.New("MongoDB client not initialized yet")

}

func DeleteOne(collectionName string, filter bson.D) (int64, error) {
	if client != nil {
		collection := client.Database(dbName).Collection(collectionName)

		result, err := collection.DeleteOne(context.Background(), filter)

		if err != nil {
			return -1, err
		}

		return result.DeletedCount, nil
	}

	return -1, errors.New("MongoDB client not initialized yet")

}

func Update(collectionName string, filter bson.D, update bson.D) (*UpdateResponse, error) {
	if client != nil {
		collection := client.Database(dbName).Collection(collectionName)

		result, err := collection.UpdateMany(context.Background(), filter, update)

		if err != nil {
			return nil, err
		}

		return &UpdateResponse{
			Modified: result.ModifiedCount,
			Matched:  result.MatchedCount,
		}, nil
	}

	return nil, errors.New("MongoDB client not initialized yet")
}

func UpdateOne(collectionName string, filter bson.D, update bson.D) (*UpdateResponse, error) {
	if client != nil {
		collection := client.Database(dbName).Collection(collectionName)

		result, err := collection.UpdateOne(context.Background(), filter, update)

		if err != nil {
			return nil, err
		}

		return &UpdateResponse{
			Modified: result.ModifiedCount,
			Matched:  result.MatchedCount,
		}, nil
	}

	return nil, errors.New("MongoDB client not initialized yet")
}

func InsertOne(collectionName string, object interface{}) *InsertResponse {
	if client != nil {
		collection := client.Database(dbName).Collection(collectionName)

		result, err := collection.InsertOne(context.Background(), object)

		if err != nil {
			return &InsertResponse{"", err}
		}

		return &InsertResponse{result.InsertedID.(primitive.ObjectID).Hex(), err}
	}

	return &InsertResponse{"", errors.New("MongoDB client not initialized yet")}
}

func FindOne(collectionName string, filter bson.D, opts *options.FindOneOptions) (bson.M, error) {
	if client != nil {
		if opts == nil {
			opts = &options.FindOneOptions{}
		}
		var result bson.M

		collection := client.Database(dbName).Collection(collectionName)

		if err := collection.FindOne(context.Background(), filter, opts).Decode(&result); err != nil {
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

func Find(collectionName string, filter bson.D, opts *options.FindOptions) ([]bson.M, error) {
	if client != nil {
		if opts == nil {
			opts = &options.FindOptions{}
		}

		var result []bson.M

		collection := client.Database(dbName).Collection(collectionName)

		cursor, err := collection.Find(context.Background(), filter, opts)

		if err != nil {
			if err == mongo.ErrNoDocuments {
				return nil, nil
			} else {
				return nil, err
			}
		}

		if err := cursor.All(context.Background(), &result); err != nil {
			return nil, err
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
