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
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net/http"
	"strconv"
	"time"
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

func insertImage(image model.Image, channel chan *database.InsertResponse) {
	res := database.InsertOne("images", image)
	channel <- res
}

func getFindOneImages(filter bson.D, opts *options.FindOneOptions, channel chan *database.FindOneResponse) {
	res := database.FindOne("images", filter, opts)
	channel <- res
}

func getFindImages(filter bson.D, opts *options.FindOptions, channel chan *database.FindResponse) {
	res := database.Find("images", filter, opts)
	channel <- res
}

func validateAcceptableMIMEType(mimeType string) bool {
	return mimeset[mimeType]
}

/**
Builds the image query based on the parameters passed in the request.

Returns a BSON Document representing the database query to be built, and a number that represents the limit.
*/
func buildImageQuery(r *http.Request) (bson.D, int64) {
	loggedInUser := getUserIDFromTokenNotStrictValidation(r)

	before := time.Time{}
	after := time.Time{}
	var limit int64 = 10
	user := ""

	if beforeQuery, beforeOK := r.URL.Query()["before"]; beforeOK && len(beforeQuery) > 0 && len(beforeQuery[0]) > 0 {
		if conv, convErr := strconv.ParseInt(beforeQuery[0], 10, 64); convErr == nil {
			before = time.Unix(conv, 0)
		}
	}
	if afterQuery, afterOK := r.URL.Query()["after"]; afterOK && len(afterQuery) > 0 && len(afterQuery[0]) > 0 {
		if conv, convErr := strconv.ParseInt(afterQuery[0], 10, 64); convErr == nil {
			after = time.Unix(conv, 0)
		}
	}
	if limitQuery, limitOK := r.URL.Query()["limit"]; limitOK && len(limitQuery) > 0 && len(limitQuery[0]) > 0 {
		if conv, convErr := strconv.ParseInt(limitQuery[0], 10, 64); convErr == nil {
			limit = conv
		}
	}
	if userQuery, userOK := r.URL.Query()["user"]; userOK && len(userQuery) > 0 && len(userQuery[0]) > 0 {
		user = userQuery[0]
	}

	var subFilters []interface{}

	if !before.IsZero() {
		subFilters = append(subFilters, bson.D{{"uploadDateTime", bson.D{{"$lte", primitive.NewDateTimeFromTime(before)}}}})
	}

	if !after.IsZero() {
		subFilters = append(subFilters, bson.D{{"uploadDateTime", bson.D{{"$gte", primitive.NewDateTimeFromTime(after)}}}})
	}

	visibilityFilters := []interface{}{bson.D{{"accessLevel", "public"}}}

	if loggedInUser != "" {
		visibilityFilters = append(visibilityFilters, bson.D{{"accessListIDs", loggedInUser}})
		visibilityFilters = append(visibilityFilters, bson.D{{"authorid", loggedInUser}})
	}

	if user != "" {
		subFilters = append(subFilters, bson.D{{"$and", []interface{}{
			bson.D{{"$or", visibilityFilters}},
			bson.D{{"authorid", user}},
		}}})
	} else {
		subFilters = append(subFilters, bson.D{{"$or", visibilityFilters}})
	}

	return bson.D{{"$and", subFilters}}, limit
}

/**
	Gets user ID from the token.

	This is required for endpoints which might require a user ID, but are not checked by the middleware because they
	are not strictly necessary. JWT Middleware is only run on functions where the user MUST be authenticated.
 */
func getUserIDFromTokenNotStrictValidation(r *http.Request) string {
	cookie, err := r.Cookie("token")
	if err == nil {
		token := cookie.Value
		valid, vErr := routes.VerifyJWT(token)

		if valid && vErr == nil {
			return getUserIDFromToken(r)
		}
	}

	return ""
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
				accessListIDsString := r.FormValue("accessListIDs")
				caption := r.FormValue("caption")

				var accessListIDs []string

				jsonParseErr := json.Unmarshal([]byte(accessListIDsString), &accessListIDs)

				if jsonParseErr != nil {
					log.Println("passed empty access List IDs")
					accessListIDs = []string{}
				}
				image := model.Image{
					AuthorID:       authorID,
					AccessLevel:    accessLevel,
					Caption:		caption,
					UploadDate:		time.Now(),
					AccessListIDs:  accessListIDs,
					Likes:          []string{},
				}

				channel := make(chan *database.InsertResponse)

				go insertImage(image, channel)

				insertResponse := <-channel

				if insertResponse.Err != nil {
					sendInternalServerError(w)
				} else {
					id := insertResponse.ID

					_, uploadErr := s3Uploader.Upload(&s3manager.UploadInput{
						Bucket: aws.String(bucketName),
						Key:    aws.String(id),
						Body:   file,
					})

					if uploadErr != nil {
						sendInternalServerError(w)
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

func editImageACL(w http.ResponseWriter, r *http.Request) {

}

func likeImage(w http.ResponseWriter, r *http.Request) {

}

func unlikeImage(w http.ResponseWriter, r *http.Request) {

}

func getImage(w http.ResponseWriter, r *http.Request) {

}

func deleteImage(w http.ResponseWriter, r *http.Request) {

}

func deleteAllImages(w http.ResponseWriter, r *http.Request) {

}

/**
	Gets the metadata (not the actual image files) of the images in the database based on the queries passed in, in chronologically descending order.

	Accepted query parameters:
		- before: UNIX time stamp representing the latest image that can be uploaded.
		- after: UNIX time stamp repesenting the earliest image that should be fetched.
		- limit: integer. the limit on the number of images to fetch. Default 10 if not specified.
		- user: userID. Gets images from a particular user.
 */
func getImagesMetadata(w http.ResponseWriter, r *http.Request) {
	filter, limit := buildImageQuery(r)

	opts := &options.FindOptions{
		Limit:               &limit,
		Sort:                bson.D{{"uploadDateTime", -1}},
	}

	channel := make(chan *database.FindResponse)

	go getFindImages(filter, opts, channel)

	res := <- channel

	imageList := []model.Image{}

	if res.Err != nil {
		sendInternalServerError(w)
		return
	}

	for i:=0; i < len(res.Result); i++ {
		image := model.Image{}

		bsonBytes, _ := bson.Marshal(res.Result[i])

		_ = bson.Unmarshal(bsonBytes, &image)

		imageList = append(imageList, image)
	}

	marshalled, _ := json.Marshal(imageList)

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(200)
	_, _ = w.Write(marshalled)
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
	r.HandleFunc("/getImagesMetadata", getImagesMetadata).Methods("GET")

	s := r.Methods("PUT", "DELETE", "PATCH").Subrouter()

	s.Use(routes.JWTMiddleware)

	s.HandleFunc("/addImage", addNewImage).Methods("PUT")
	s.HandleFunc("/deleteImage", deleteImage).Methods("DELETE")
	s.HandleFunc("/editImageACL", editImageACL).Methods("PATCH")
	s.HandleFunc("/likeImage", likeImage).Methods("PUT")
	s.HandleFunc("/unlikeImage", unlikeImage).Methods("DELETE")
	s.HandleFunc("/deleteAllImages", deleteAllImages).Methods("DELETE")
}
