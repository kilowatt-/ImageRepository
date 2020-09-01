package routes

import (
	"github.com/gorilla/mux"
	routes "github.com/kilowatt-/ImageRepository/routes/middleware"
	"net/http"
)

func addNewImage(w http.ResponseWriter, r *http.Request) {

}

func deleteImage(w http.ResponseWriter, r *http.Request) {

}

func editImageACL(w http.ResponseWriter, r *http.Request) {

}

func likeImage(w http.ResponseWriter, r *http.Request) {

}

func unlikeImage(w http.ResponseWriter, r *http.Request) {

}

func favouriteImage(w http.ResponseWriter, r *http.Request) {

}

func unfavouriteImage(w http.ResponseWriter, r *http.Request) {

}

func getImage(w http.ResponseWriter, r *http.Request) {

}

func getImagesFromUser(w http.ResponseWriter, r *http.Request) {

}

func deleteAllImages(w http.ResponseWriter, r *http.Request) {

}

func getLatestImages(w http.ResponseWriter, r *http.Request) {

}

func getTopImages(w http.ResponseWriter, r *http.Request) {

}

func serveImageRoutes(r *mux.Router) {

	r.HandleFunc("/getImage/:id", getImage).Methods("GET")
	r.HandleFunc("/getImagesFrom/:user", getImagesFromUser).Methods("GET")
	r.HandleFunc("/getLatestImages", getLatestImages).Methods("GET")
	r.HandleFunc("/getTopImages", getTopImages).Methods("GET")

	s := r.Methods("PUT", "DELETE", "PATCH").Subrouter()

	s.Use(routes.JWTMiddleware)

	s.HandleFunc("/addImage", addNewImage).Methods("PUT")
	s.HandleFunc("/deleteImage", deleteImage).Methods("DELETE")
	s.HandleFunc("/editImageACL", editImageACL).Methods("PATCH")
	s.HandleFunc("/likeImage", likeImage).Methods("PUT")
	s.HandleFunc("/unlikeImage", unlikeImage).Methods("DELETE")
	s.HandleFunc("/favouriteImage", favouriteImage).Methods("PUT")
	s.HandleFunc("/unfavouriteImage", unfavouriteImage).Methods("DELETE")
	s.HandleFunc("/deleteAllImages", deleteAllImages).Methods("DELETE")
}