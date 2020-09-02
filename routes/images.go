package routes

import (
	"encoding/json"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/kilowatt-/ImageRepository/database"
	"github.com/kilowatt-/ImageRepository/model"
	routes "github.com/kilowatt-/ImageRepository/routes/middleware"
	"log"
	"net/http"
)

const bucketName = "imgrepository-cdn"

var awsSession *session.Session = nil
var s3Client *s3.S3 = nil
var s3Uploader *s3manager.Uploader = nil

var mimeset = map[string]bool{
	"image/bmp":  true,
	"image/gif":  true,
	"image/jpeg": true,
	"image/png":  true,
	"image/webp": true,
}

func insertImage(image model.Image, channel chan database.InsertResponse) {
	res := database.InsertOne("images", image)
	channel <- res
}

func validateAcceptableMIMEType(mimeType string) bool {
	return mimeset[mimeType]
}

func getUserIDFromToken(r *http.Request) string {
	cookie, _ := r.Cookie("token")
	token := cookie.Value

	parsed, _, _ := new(jwt.Parser).ParseUnverified(token, &jwt.MapClaims{})

	claims := parsed.Claims.(*jwt.MapClaims)

	return (*claims)["id"].(string)
}

func addNewImage(w http.ResponseWriter, r *http.Request) {

	parseFormErr := r.ParseMultipartForm(10 << 20)
	// Maximum total form data size: 10MB.
	if parseFormErr != nil {
		http.Error(w, parseFormErr.Error(), http.StatusBadRequest)
	} else {
		file, handler, formFileErr := r.FormFile("file")

		if formFileErr != nil {
			http.Error(w, "Error parsing file", http.StatusBadRequest)
		} else {
			mimeType := handler.Header

			if !validateAcceptableMIMEType(mimeType.Get("Content-Type")) {
				http.Error(w, "Uploaded non-image file type", http.StatusBadRequest)
			} else {
				authorID := getUserIDFromToken(r)
				accessLevel := r.FormValue("accessLevel")
				accessListType := r.FormValue("accessListType")
				accessListIDsString := r.FormValue("accessListIDs")

				var accessListIDs []string

				jsonParseErr := json.Unmarshal([]byte(accessListIDsString), &accessListIDs)

				if jsonParseErr != nil {
					log.Println("passed empty access List IDs")
					accessListIDs = []string{}
				}
				image := model.Image{
					AuthorID:       authorID,
					AccessLevel:    accessLevel,
					AccessListType: accessListType,
					AccessListIDs:  accessListIDs,
					Likes:          0,
				}

				channel := make(chan database.InsertResponse)

				go insertImage(image, channel)

				insertResponse := <-channel

				if insertResponse.Err != nil {
					http.Error(w, "Internal server error", http.StatusInternalServerError)
				} else {
					id := insertResponse.ID

					_, uploadErr := s3Uploader.Upload(&s3manager.UploadInput{
						Bucket: aws.String(bucketName),
						Key:    aws.String(id),
						Body:   file,
					})

					if uploadErr != nil {
						http.Error(w, "Internal server error", http.StatusInternalServerError)
					} else {
						insertResponse.Err = nil
						w.WriteHeader(200)
						jsonResponse, _ := json.Marshal(insertResponse)
						_, _ = w.Write(jsonResponse)
					}
				}
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
func initAWS() {
	awsSession = session.Must(session.NewSessionWithOptions(session.Options{SharedConfigState: session.SharedConfigEnable}))
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
