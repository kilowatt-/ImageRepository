package routes

import (
	"github.com/kilowatt-/ImageRepository/database"
	"github.com/kilowatt-/ImageRepository/model"
	"net/http"
)

func createUser(user model.User, errChan chan error) {
	_, err := database.InsertOne("users", user)

	errChan <- err
}


func verifyJWT() bool {
	return false
}


func handleCreateUser(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		errChan := make(chan error)
		go createUser(
			model.User{"Kean", "hello", "world"},
			errChan,
		)
		createUserErr := <- errChan

		if createUserErr != nil {

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
	http.HandleFunc("/api/users/signup", handleCreateUser)
	http.HandleFunc("/api/users/login", handleLogin)
}
