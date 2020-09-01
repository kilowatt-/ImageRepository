package routes

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/gorilla/mux"
	"github.com/kilowatt-/ImageRepository/database"
	"github.com/kilowatt-/ImageRepository/model"
	routes "github.com/kilowatt-/ImageRepository/routes/middleware"
	"net/http"
)

const bucketName = "imgrepository-cdn"

var awsSession *session.Session = nil
var s3Client *s3.S3 = nil
var s3Uploader *s3manager.Uploader = nil

func insertImage(image model.Image, channel chan database.InsertResponse) {
	res := database.InsertOne("images", image)
	channel <- res
}

func validateImageIsAccepted(base64 string) bool {
	return false
}

func addNewImage(w http.ResponseWriter, r *http.Request) {
	var image model.Image

	err := json.NewDecoder(r.Body).Decode(&image)

	if err != nil {
		fmt.Println(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	} else {
		base64 := image.Base64

		if !validateImageIsAccepted(base64) {
			http.Error(w, "Image format not accepted", http.StatusBadRequest)
		} else {
			image.Base64 = ""

			channel := make(chan database.InsertResponse)

			go insertImage(image, channel)

			insertResponse := <- channel

			if insertResponse.Err != nil {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			} else {


			}
		}
	}
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

/**
	Initializes the AWS session, S3 Client and S3 Uploader.
 */
func initAWS()  {
	awsSession = session.Must(session.NewSessionWithOptions(session.Options{ SharedConfigState: session.SharedConfigEnable }))
	s3Client = s3.New(awsSession)
	s3Uploader = s3manager.NewUploader(awsSession)
}


func serveImageRoutes(r *mux.Router) {
	initAWS()

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