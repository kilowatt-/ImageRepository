package users

import (
	"errors"
	"github.com/kilowatt-/ImageRepository/database"
	"github.com/kilowatt-/ImageRepository/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

const UserNotFound = "user not found"
const PasswordNotMatching = "password does not match"
const InvalidHex = "invalid hex"

type FindUserResponse struct {
	User model.User
	Err  error
}

func createUser(user model.User, channel chan *database.InsertResponse) {
	res := database.InsertOne("users", user, nil)

	channel <- res
}

/**
Gets users from the database according to the given filter and projection.
*/
func GetUsersFromDatabase(filter interface{}, projection interface{}, channel chan []FindUserResponse, optionsList ...*options.FindOptions) {
	opts := &options.FindOptions{}

	if projection != nil {
		opts.SetProjection(projection)
	}

	if len(optionsList) > 0 {
		for _, o := range optionsList {
			opts = options.MergeFindOptions(opts, o)
		}
	}

	var userArray = []FindUserResponse{}

	res := database.Find("users", filter, opts)

	if res.Err != nil {
		userArray = append(userArray, FindUserResponse{model.User{}, res.Err})
	} else {
		for i := 0; i < len(res.Result); i++ {
			var user model.User
			bsonBytes, _ := bson.Marshal(res.Result[i])
			_ = bson.Unmarshal(bsonBytes, &user)

			userArray = append(userArray, FindUserResponse{user, nil})
		}
	}

	channel <- userArray
}

/**
Gets user from database based on email and password.

If user is not found, an error will be returned.
*/
func GetUserWithLogin(email string, password string, channel chan FindUserResponse) {
	filter := bson.D{{"email", email}}

	user := model.User{}

	response := database.FindOne("users", filter, nil)

	if response.Err != nil {
		channel <- FindUserResponse{User: user, Err: response.Err}
	} else if len(response.Result) == 0 {
		channel <- FindUserResponse{User: user, Err: errors.New(UserNotFound)}
	} else {

		bsonBytes, _ := bson.Marshal(response.Result)

		_ = bson.Unmarshal(bsonBytes, &user)

		pwdErr := bcrypt.CompareHashAndPassword(user.Password, []byte(password))

		if pwdErr != nil {
			channel <- FindUserResponse{User: user, Err: errors.New(PasswordNotMatching)}
		} else {
			channel <- FindUserResponse{user, nil}
		}
	}
}

