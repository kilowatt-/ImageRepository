package routes

import (
	"encoding/json"
	"errors"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/kilowatt-/ImageRepository/database"
	"github.com/kilowatt-/ImageRepository/model"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)


const UserNotFound = "user not found"
const PasswordNotMatching = "password does not match"
const JWTKeyNotFound = "jwt key not found"

type createUserResponse struct {
	id string
	err error
}

type findUserResponse struct {
	user model.User
	err error
}

func createUser(user model.User, channel chan createUserResponse) {
	id, err := database.InsertOne("users", user)

	channel <- createUserResponse{id, err}
}

/**
	Gets user from database based on email and password.

	If user is not found, an error will be returned.
 */
func getUser(email string, password string, channel chan findUserResponse) {
	filter := bson.D{{"email",  email}}

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

/**
	Creates a login token that is a JWT. Expires 1 hour after creation.

	Returns the signed token, expiry date, and error.
 */
func createLoginToken(user model.User) (string, time.Time, error) {
	secretKey, keyExists := os.LookupEnv("JWT_KEY")

	if !keyExists {
		return "", time.Now(), errors.New(JWTKeyNotFound)
	}

	now := time.Now()
	expiry := now.Add(time.Hour * 1)

	claims := jwt.MapClaims{}

	claims["authorized"] = true
	claims["id"] = user.ID
	claims["email"] = user.Email
	claims["name"] = user.Name
	claims["loginTime"] = now.Unix()
	claims["expiry"] = expiry.Unix()

	unsignedJwt := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	token, err := unsignedJwt.SignedString([]byte(secretKey))

	if err != nil {
		return "", time.Now(), err
	}

	return token, expiry, nil
}

func verifyJWT(jwt string) bool {
	return false
}

func verifyEmail(email string) bool {
	matched, err := regexp.MatchString("(?:[a-z0-9!#$%&'*+/=?^_`{|}~-]+(?:\\.[a-z0-9!#$%&'*+/=?^_`{|}~-]+)*|\"(?:[\\x01-\\x08\\x0b\\x0c\\x0e-\\x1f\\x21\\x23-\\x5b\\x5d-\\x7f]|\\\\[\\x01-\\x09\\x0b\\x0c\\x0e-\\x7f])*\")@(?:(?:[a-z0-9](?:[a-z0-9-]*[a-z0-9])?\\.)+[a-z0-9](?:[a-z0-9-]*[a-z0-9])?|\\[(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?|[a-z0-9-]*[a-z0-9]:(?:[\\x01-\\x08\\x0b\\x0c\\x0e-\\x1f\\x21-\\x5a\\x53-\\x7f]|\\\\[\\x01-\\x09\\x0b\\x0c\\x0e-\\x7f])+)\\])", email)
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

	for i:=0; i < pwdLength; i++ {
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

func handleSignUp(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		if parseFormErr := r.ParseForm(); parseFormErr != nil {
			http.Error(w, "Sent invalid form", 400)
		}

		name := r.FormValue("name")
		email := r.FormValue("email")
		password := r.FormValue("password")

		if !verifyEmail(email) {
			http.Error(w, "Invalid email", http.StatusBadRequest)
			return
		}

		if !verifyPassword(password) {
			http.Error(w, "Password does not meet complexity requirements", http.StatusBadRequest)
			return
		}

		hashed, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

		urChannel := make(chan createUserResponse)
		go createUser(
			model.User{Name: name, Email: email, Password: hashed},
			urChannel,
		)
		createdUser := <- urChannel

		if createdUser.err != nil {
			log.Println(createdUser.err)

			if strings.Contains(createdUser.err.Error(), "E11000") {
				http.Error(w, "Email " + email + " already registered", http.StatusConflict)
			} else {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}

		} else {
			log.Println("Created user with ID " + createdUser.id)
			w.WriteHeader(http.StatusOK)
			_, wError := w.Write([]byte("Created user with ID " + createdUser.id))

			if wError != nil {
				log.Println("Error while writing: " + wError.Error())
			}
		}

	default:
		http.Error(w, "Only POST requests are supported on this endpoint", http.StatusNotFound)
	}
}

type loginResponse struct {
	Token string `json:"token,omitempty"`
	Expiry time.Time `json:"expiry,omitempty"`
	User model.User `json:"user,omitempty"`
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch r.Method {
	case "POST":
		parseFormErr := r.ParseForm()

		if parseFormErr != nil {
			http.Error(w, "Sent invalid form", 400)
		} else {
			email := r.FormValue("email")
			password := r.FormValue("password")

			channel := make(chan findUserResponse)
			go getUser(email, password, channel)

			res := <- channel

			if res.err != nil {
				if res.err.Error() == UserNotFound || res.err.Error() == PasswordNotMatching {
					w.WriteHeader(http.StatusNotFound)
					_, _ = w.Write([]byte("Provided email/password do not match"))
				} else {
					http.Error(w, "Internal server error", http.StatusInternalServerError)
				}
			} else {
				token, expiry, jwtErr := createLoginToken(res.user)

				if jwtErr != nil {
					log.Println(jwtErr)
					http.Error(w, "Internal server error", http.StatusInternalServerError)
				} else {
					res.user.Password = nil
					jsonResponse, jsonErr := json.Marshal(loginResponse{
						Token: token,
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



	default:
		http.Error(w, "Only POST requests are supported on this endpoint", http.StatusNotFound)
	}
}

func serveUserRoutes(r *mux.Router) {
	r.HandleFunc("/api/users/signup", handleSignUp)
	r.HandleFunc("/api/users/login", handleLogin)
}
