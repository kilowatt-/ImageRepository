package routes

import (
	"github.com/gorilla/mux"
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
	r.HandleFunc("/api/images/addImage", addNewImage).Methods("PUT")
	r.HandleFunc("/api/images/deleteImage", deleteImage).Methods("DELETE")
	r.HandleFunc("/api/images/editImageACL", editImageACL).Methods("PATCH")
	r.HandleFunc("/api/images/likeImage", likeImage).Methods("PUT")
	r.HandleFunc("/api/images/unlikeImage", unlikeImage).Methods("DELETE")
	r.HandleFunc("/api/images/favouriteImage", favouriteImage).Methods("PUT")
	r.HandleFunc("/api/images/unfavouriteImage", unfavouriteImage).Methods("DELETE")
	r.HandleFunc("/api/images/getImage/:id", getImage).Methods("GET")
	r.HandleFunc("/api/images/getImagesFrom/:user", getImagesFromUser).Methods("GET")
	r.HandleFunc("/api/images/deleteAllImages", deleteAllImages).Methods("DELETE")
	r.HandleFunc("/api/images/getLatestImages", getLatestImages).Methods("GET")
	r.HandleFunc("/api/images/getTopImages", getTopImages).Methods("GET")
}