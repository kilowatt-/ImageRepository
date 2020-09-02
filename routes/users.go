package routes

import (
	"encoding/json"
	"errors"
	"github.com/gorilla/mux"
	"github.com/kilowatt-/ImageRepository/database"
	"github.com/kilowatt-/ImageRepository/model"
	routes "github.com/kilowatt-/ImageRepository/routes/middleware"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"
)

const UserNotFound = "user not found"
const PasswordNotMatching = "password does not match"


type findUserResponse struct {
	user model.User
	err  error
}

func createUser(user model.User, channel chan database.InsertResponse) {
	res := database.InsertOne("users", user)

	channel <- res
}


/**
	Gets user from database based on email and password.

	If user is not found, an error will be returned.
*/
func getUserWithLogin(email string, password string, channel chan findUserResponse) {
	filter := bson.D{{"email", email}}

	blankUser := model.User{
		Name:     "",
		Email:    "",
		Password: nil,
	}

	response, err := database.FindOne("users", filter)

	if err != nil {
		channel <- findUserResponse{user: blankUser, err: err}
	} else if len(response) == 0 {
		channel <- findUserResponse{user: blankUser, err: errors.New(UserNotFound)}
	} else {
		var user model.User

		bsonBytes, _ := bson.Marshal(response)

		_ = bson.Unmarshal(bsonBytes, &user)

		pwdErr := bcrypt.CompareHashAndPassword(user.Password, []byte(password))

		if pwdErr != nil {
			channel <- findUserResponse{user: blankUser, err: errors.New(PasswordNotMatching)}
		} else {
			channel <- findUserResponse{user, nil}
		}
	}
}

func verifyEmail(email string) bool {
	matched, err := regexp.MatchString("(?:[a-z0-9!#$%&'*+/=?^_`{|}~-]+(?:\\.[a-z0-9!#$%&'*+/=?^_`{|}~-]+)*|\"(?:[\\x01-\\x08\\x0b\\x0c\\x0e-\\x1f\\x21\\x23-\\x5b\\x5d-\\x7f]|\\\\[\\x01-\\x09\\x0b\\x0c\\x0e-\\x7f])*\")@(?:(?:[a-z0-9](?:[a-z0-9-]*[a-z0-9])?\\.)+[a-z0-9](?:[a-z0-9-]*[a-z0-9])?|\\[(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?|[a-z0-9-]*[a-z0-9]:(?:[\\x01-\\x08\\x0b\\x0c\\x0e-\\x1f\\x21-\\x5a\\x53-\\x7f]|\\\\[\\x01-\\x09\\x0b\\x0c\\x0e-\\x7f])+)\\])", email)
	return err == nil && matched
}

func verifyUserHandle(userHandle string) bool {
	matched, err := regexp.MatchString("^[a-zA-Z0-9]+$", userHandle)
	return err == nil && matched
}

/**
Verifies if password meets the following criteria:
- At least 8 characters
- 1 Uppercase
- 1 Lowercase
- 1 Digit
*/
func verifyPassword(password string) bool {
	hasDigit, hasUppercase, hasLowerCase := false, false, false
	pwdLength := len(password)

	if pwdLength < 8 {
		return false
	}

	for i := 0; i < pwdLength; i++ {
		character := password[i]

		if character >= 'a' && character <= 'z' {
			hasLowerCase = true
		} else if character >= 'A' && character <= 'Z' {
			hasUppercase = true
		} else if character >= '0' && character <= '9' {
			hasDigit = true
		}
	}

	return hasDigit && hasUppercase && hasLowerCase
}

/**
	Handles sign up. Takes in a name, email and password, verifies the inputs, and creates the users.

	Returns:
	- 200: User was created. Returns ID.
	- 400: If any complexity requirement was not met, or an invalid form was sent.
	- 409: Conflict, if the user was already registered.
	- 500: If there is an error on the server (Database error, etc).
 */
func handleSignUp(w http.ResponseWriter, r *http.Request) {
	if parseFormErr := r.ParseForm(); parseFormErr != nil {
		http.Error(w, "Sent invalid form", 400)
	}

	name := r.FormValue("name")
	userHandle := r.FormValue("userHandle")
	email := r.FormValue("email")
	password := r.FormValue("password")

	if !verifyUserHandle(userHandle) {
		http.Error(w, "Invalid userhandle", http.StatusBadRequest)
	}

	if !verifyEmail(email) {
		http.Error(w, "Invalid email", http.StatusBadRequest)
		return
	}

	if !verifyPassword(password) {
		http.Error(w, "Password does not meet complexity requirements", http.StatusBadRequest)
		return
	}

	hashed, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	urChannel := make(chan database.InsertResponse)
	go createUser(
		model.User{Name: name, UserHandle: userHandle,Email: email, Password: hashed},
		urChannel,
	)
	createdUser := <-urChannel

	if createdUser.Err != nil {
		log.Println(createdUser.Err)

		if strings.Contains(createdUser.Err.Error(), "E11000") {
			if (strings.Contains(createdUser.Err.Error(), "index: userHandle_1")) {
				http.Error(w, "Userhandle "+userHandle+" already registered", http.StatusConflict)
			} else {
				http.Error(w, "Email "+email+" already registered", http.StatusConflict)
			}
		} else {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}

	} else {
		log.Println("Created user with ID " + createdUser.ID)
		w.WriteHeader(http.StatusOK)
		_, wError := w.Write([]byte("Created user with ID " + createdUser.ID))

		if wError != nil {
			log.Println("Error while writing: " + wError.Error())
		}
	}

}

type loginResponse struct {
	Token  string     `json:"token,omitempty"`
	Expiry time.Time  `json:"expiry,omitempty"`
	User   model.User `json:"user,omitempty"`
}

/**
	Handles a login request.

	Returns:
	- 200: OK, if the username and password match. Will return a JWT along with the expiry date and userinfo.
	- 404: If the user was not found, or the username and password don't match.
	- 500: If there is an internal server error.
 */
func handleLogin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	parseFormErr := r.ParseForm()

	if parseFormErr != nil {
		http.Error(w, "Sent invalid form", http.StatusBadRequest)
	} else {
		email := r.FormValue("email")
		password := r.FormValue("password")

		channel := make(chan findUserResponse)
		go getUserWithLogin(email, password, channel)

		res := <-channel

		if res.err != nil {
			if res.err.Error() == UserNotFound || res.err.Error() == PasswordNotMatching {
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write([]byte("Provided email/password do not match"))
			} else {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
		} else {
			token, expiry, jwtErr := routes.CreateLoginToken(res.user)

			if jwtErr != nil {
				log.Println(jwtErr)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			} else {
				res.user.Password = nil
				jsonResponse, jsonErr := json.Marshal(loginResponse{
					Token:  token,
					Expiry: expiry,
					User:   res.user,
				})

				if jsonErr != nil {
					log.Println(jsonErr)
					http.Error(w, "Internal server error", http.StatusInternalServerError)
				} else {
					w.WriteHeader(200)
					w.Write(jsonResponse)
				}
			}
		}
	}
}

func serveUserRoutes(r *mux.Router) {
	r.HandleFunc("/signup", handleSignUp).Methods("POST")
	r.HandleFunc("/login", handleLogin).Methods("POST")
}
