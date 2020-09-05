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
	Err error
}

type FindOneResponse struct {
	Result bson.M
	Err error
}

type FindResponse struct {
	Result []bson.M
	Err error
}

type DeleteResponse struct {
	NumberDeleted int64
	Err error
}

func Delete(collectionName string, filter interface{}, opts *options.DeleteOptions) *DeleteResponse {
	if client != nil {
		if opts == nil {
			opts = &options.DeleteOptions{}
		}

		collection := client.Database(dbName).Collection(collectionName)

		result, err := collection.DeleteMany(context.Background(), filter)

		if err != nil {
			return &DeleteResponse{-1, err}
		}

		return &DeleteResponse{result.DeletedCount, nil}
	}

	return &DeleteResponse{-1, errors.New("MongoDB client not initialized yet")}

}

func DeleteOne(collectionName string, filter interface{}, opts *options.DeleteOptions) *DeleteResponse {
	if client != nil {

		if opts == nil {
			opts = &options.DeleteOptions{}
		}

		collection := client.Database(dbName).Collection(collectionName)

		result, err := collection.DeleteOne(context.Background(), filter)

		if err != nil {
			return &DeleteResponse{-1, err}
		}

		return &DeleteResponse{result.DeletedCount, nil}
	}

	return &DeleteResponse{-1, errors.New("MongoDB client not initialized yet")}
}

func Update(collectionName string, filter interface{}, update bson.D, opts *options.UpdateOptions) (*UpdateResponse) {
	if client != nil {
		if opts == nil {
			opts = &options.UpdateOptions{}
		}
		collection := client.Database(dbName).Collection(collectionName)

		result, err := collection.UpdateMany(context.Background(), filter, update, opts)

		if err != nil {
			return &UpdateResponse{0, 0, err}
		}

		return &UpdateResponse{
			Modified: result.ModifiedCount,
			Matched:  result.MatchedCount,
			Err: nil,
		}
	}

	return &UpdateResponse{Modified: 0, Matched: 0, Err: errors.New("MongoDB client not initialized yet")}
}

func UpdateOne(collectionName string, filter interface{}, update bson.D, opts *options.UpdateOptions) *UpdateResponse {
	if client != nil {
		if opts == nil {
			opts = &options.UpdateOptions{}
		}
		collection := client.Database(dbName).Collection(collectionName)

		result, err := collection.UpdateOne(context.Background(), filter, update, opts)

		if err != nil {
			return &UpdateResponse{0, 0, err}
		}

		return &UpdateResponse{
			Modified: result.ModifiedCount,
			Matched:  result.MatchedCount,
			Err: nil,
		}
	}

	return &UpdateResponse{Modified: 0, Matched: 0, Err: errors.New("MongoDB client not initialized yet")}
}

func InsertOne(collectionName string, object interface{}, opts *options.InsertOneOptions) *InsertResponse {
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

func FindOne(collectionName string, filter interface{}, opts *options.FindOneOptions) *FindOneResponse {
	if client != nil {
		if opts == nil {
			opts = &options.FindOneOptions{}
		}
		var result bson.M

		collection := client.Database(dbName).Collection(collectionName)

		if err := collection.FindOne(context.Background(), filter, opts).Decode(&result); err != nil {
			if err == mongo.ErrNoDocuments {
				return &FindOneResponse{nil, nil}
			} else {
				return &FindOneResponse{bson.M{}, err}
			}
		}

		return &FindOneResponse{result, nil}
	}

	return &FindOneResponse{nil, errors.New("MongoDB client not initialized yet")}
}

func Find(collectionName string, filter interface{}, opts *options.FindOptions) *FindResponse {
	if client != nil {
		if opts == nil {
			opts = &options.FindOptions{}
		}

		var result []bson.M

		collection := client.Database(dbName).Collection(collectionName)

		cursor, err := collection.Find(context.Background(), filter, opts)

		if err != nil {
			if err == mongo.ErrNoDocuments {
				return &FindResponse{result, nil}
			} else {
				return &FindResponse{nil, err}
			}
		}

		if err := cursor.All(context.Background(), &result); err != nil {
			return &FindResponse{nil, err}
		}

		return &FindResponse{result, nil}
	}

	return &FindResponse{nil, errors.New("MongoDB client not initialized yet")}
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
