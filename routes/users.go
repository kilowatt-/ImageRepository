package routes

import (
	"github.com/kilowatt-/ImageRepository/database"
	"github.com/kilowatt-/ImageRepository/model"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	"regexp"
	"strings"
)

type userResponse struct {
	id string
	err error
}

func createUser(user model.User, errChan chan userResponse) {
	id, err := database.InsertOne("users", user)

	errChan <- userResponse{id, err}
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

		urChannel := make(chan userResponse)
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

func handleLogin(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
	default:
		http.Error(w, "Only POST requests are supported on this endpoint", http.StatusNotFound)
	}
}

func serveUserRoutes() {
	http.HandleFunc("/api/users/signup", handleSignUp)
	http.HandleFunc("/api/users/login", handleLogin)
}
