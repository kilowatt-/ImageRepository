package routes

import (
	"bytes"
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
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const bucketName = "imgrepository-cdn"
const invalidImageId = "invalid image id"
const imageNotFound = "image not found"

type imageDatabaseResponse struct {
	images    []*model.Image
	userIDMap *map[string]bool
	err       error
}

var publicFilter = bson.D{{"accessLevel", "public"}}
var awsSession *session.Session = nil
var s3Uploader *s3manager.Uploader = nil
var s3Downloader *s3manager.Downloader = nil

var mimeSet = map[string]bool{
	"image/bmp":  true,
	"image/gif":  true,
	"image/jpeg": true,
	"image/png":  true,
	"image/webp": true,
}

func insertImage(image model.Image, channel chan *database.InsertResponse) {
	res := database.InsertOne("images", image, nil)
	channel <- res
}

func getOneImage(filter bson.D, opts *options.FindOneOptions, channel chan imageDatabaseResponse) {
	res := database.FindOne("images", filter, opts)

	if res.Err != nil {
		channel <- imageDatabaseResponse{images: nil, err: res.Err}
	} else {
		imageList := []*model.Image{}

		if res.Result != nil {
			image := model.Image{}

			bsonBytes, _ := bson.Marshal(res.Result)

			_ = bson.Unmarshal(bsonBytes, &image)

			imageList = append(imageList, &image)
		}

		channel <- imageDatabaseResponse{images: imageList, err: nil}
	}
}

func like(userid string, imageid primitive.ObjectID, channel chan *database.UpdateResponse) {
	filter := bson.D{{"$and", []bson.D{
		{{"_id", imageid}},
		{{"$or", buildVisibilityFilters(userid)}},
	}}}

	update := bson.D{{"$addToSet", bson.D{{"likes", userid}}}}

	channel <- database.UpdateOne("images", filter, update, nil)
}

func unlike(userid string, imageid primitive.ObjectID, channel chan *database.UpdateResponse) {
	filter := bson.D{{"$and", []bson.D{
		{{"_id", imageid}},
		{{"$or", buildVisibilityFilters(userid)}},
	}}}

	update := bson.D{{"$pull", bson.D{{"likes", userid}}}}

	channel <- database.UpdateOne("images", filter, update, nil)
}

func getImagesMetadataFromDatabase(filter bson.D, opts *options.FindOptions, channel chan imageDatabaseResponse) {
	res := database.Find("images", filter, opts)

	imageList := []*model.Image{}
	var authorIDMap = make(map[string]bool)

	if res.Err != nil {
		channel <- imageDatabaseResponse{
			images:    nil,
			userIDMap: nil,
			err:       res.Err,
		}
		return
	} else {
		for i := 0; i < len(res.Result); i++ {
			image := model.Image{}

			bsonBytes, _ := bson.Marshal(res.Result[i])

			_ = bson.Unmarshal(bsonBytes, &image)

			authorIDMap[image.AuthorID] = true
			imageList = append(imageList, &image)
		}
	}

	channel <- imageDatabaseResponse{imageList, &authorIDMap, nil}
}

func validateAcceptableMIMEType(mimeType string) bool {
	return mimeSet[mimeType]
}

func buildVisibilityFilters(userid string) []interface{} {
	filters := []interface{}{publicFilter}

	if userid != "" {
		filters = append(filters, bson.D{{"accessListIDs", userid}})
		filters = append(filters, bson.D{{"authorid", userid}})
	}

	return filters
}

// Appends author to each image in the list.
func appendAuthorsToImages(images []*model.Image, idMap map[string]bool) {
	userIDs := make([]primitive.ObjectID, len(idMap))

	i := 0

	for k := range idMap {
		hex, _ := primitive.ObjectIDFromHex(k)
		userIDs[i] = hex
		i++
	}

	filter := bson.D{{"_id", bson.D{{"$in", userIDs}}}}
	projection := bson.D{{"name", 1}, {"userHandle", 1}}
	// Get author data
	c := make(chan []findUserResponse)
	go getUsersFromDatabase(filter, projection, c)

	res := <-c

	if len(res) == 1 && res[0].err != nil {
		return
	}

	userMap := make(map[string]model.User)

	for i := 0; i < len(res); i++ {
		cur := res[i].user
		userMap[cur.ID] = cur
	}

	for i := 0; i < len(images); i++ {
		cur := images[i]
		cur.SetAuthor(userMap[cur.AuthorID])
	}
}

/**
Builds the image query based on the parameters passed in the request.

Returns a BSON Document representing the database query to be built, and a number that represents the limit.
*/
func buildImageQuery(r *http.Request) (*bson.D, int64) {
	loggedInUser := getUserIDFromTokenNotStrictValidation(r)

	before := time.Time{}
	after := time.Time{}
	var limit int64 = 10
	user := []string{}

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
		user = strings.Split(userQuery[0], ",")
	}

	var subFilters []interface{}

	if !before.IsZero() {
		subFilters = append(subFilters, bson.D{{"uploadDateTime", bson.D{{"$lt", primitive.NewDateTimeFromTime(before)}}}})
	}

	if !after.IsZero() {
		subFilters = append(subFilters, bson.D{{"uploadDateTime", bson.D{{"$gt", primitive.NewDateTimeFromTime(after)}}}})
	}

	visibilityFilters := buildVisibilityFilters(loggedInUser)

	if len(user) > 0 {
		subFilters = append(subFilters, bson.D{{"$and", []interface{}{
			bson.D{{"$or", visibilityFilters}},
			bson.D{{"authorid", bson.D{{"$in", user}}}},
		}}})
	} else {
		subFilters = append(subFilters, bson.D{{"$or", visibilityFilters}})
	}

	return &bson.D{{"$and", subFilters}}, limit
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

/**
	Inserts a new image record to the database, and uploads the file to our S3 bucket.
 */
func addNewImage(w http.ResponseWriter, r *http.Request) {
	const uploadNonImageFileTypeErr = "Uploaded non-image file type"
	parseFormErr := r.ParseMultipartForm(10 << 20)
	// Maximum total form data size: 10MB.

	if parseFormErr != nil {
		http.Error(w, parseFormErr.Error(), http.StatusBadRequest)
		return
	}

	file, fileHeader, formFileErr := r.FormFile("file")

	defer file.Close()

	if formFileErr != nil {
		http.Error(w, "Error parsing file", http.StatusBadRequest)
		return
	}

	contentType := fileHeader.Header.Get("Content-Type")

	if !validateAcceptableMIMEType(contentType) {
		http.Error(w, uploadNonImageFileTypeErr, http.StatusBadRequest)
		return
	}

	buf := bytes.NewBuffer(nil)

	if _, err := io.Copy(buf, file); err != nil {
		sendInternalServerError(w)
		return
	}

	if !validateAcceptableMIMEType(http.DetectContentType(buf.Bytes())) {
		http.Error(w, uploadNonImageFileTypeErr, http.StatusBadRequest)
		return
	}

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
		AuthorID:      authorID,
		AccessLevel:   accessLevel,
		Caption:       caption,
		UploadDate:    time.Now(),
		AccessListIDs: accessListIDs,
		Likes:         []string{},
	}

	channel := make(chan *database.InsertResponse)

	go insertImage(image, channel)

	insertResponse := <-channel

	if insertResponse.Err != nil {
		sendInternalServerError(w)
		return
	}
	id := insertResponse.ID

	_, uploadErr := s3Uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(id),
		Body:   file,
	})

	if uploadErr != nil {
		sendInternalServerError(w)
		return
	}
	insertResponse.Err = nil
	w.WriteHeader(200)
	jsonResponse, _ := json.Marshal(insertResponse)
	_, _ = w.Write(jsonResponse)
}

func likeUnlikeImage(w http.ResponseWriter, r *http.Request, isLike bool) {
	uid := getUserIDFromToken(r)

	image := &model.Image{}

	err := json.NewDecoder(r.Body).Decode(&image)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if image.ID == "" {
		http.Error(w, "no image ID passed in", http.StatusBadRequest)
		return
	}

	hex, hexErr := primitive.ObjectIDFromHex(image.ID)

	if hexErr != nil {
		http.Error(w, invalidImageId, http.StatusBadRequest)
		return
	}

	log.Println(hex)

	channel := make(chan *database.UpdateResponse)

	if isLike {
		go like(uid, hex, channel)
	} else {
		go unlike(uid, hex, channel)
	}

	res := <- channel

	if res.Matched == 0 {
		http.Error(w, imageNotFound, http.StatusNotFound)
		return
	}

	if res.Modified == 0 {
		if isLike {
			http.Error(w, "already liked image", http.StatusConflict)
		} else {
			http.Error(w, "already unliked image", http.StatusConflict)
		}
		return
	}

	w.WriteHeader(200)
}

/**
	[PUT]

	Adds the selected user IDs to
 */
func editImageACL(w http.ResponseWriter, r *http.Request) {

}

/**
[PUT]

Adds this user to the image's like list.

JSON body parameters:
	- id: the image ID.

Returns:
	- 200: Image liked successfully.
	- 400: id not passed in, or invalid id passed in
	- 404: Image not found, or user not authorised to view image. (no difference.)
	- 409: User already liked image. No-op.
*/
func likeImage(w http.ResponseWriter, r *http.Request) {
	likeUnlikeImage(w, r, true)
}

/**
[DELETE]

Removes this user from the image's unlike list.

Query parameters:
	- id: the image ID.

Returns:
	- 200: Image liked successfully.
	- 400: id not passed in.
	- 404: Image not found, or user not authorised to view image. (no difference.)
	- 409: User already liked image. No-op.
*/
func unlikeImage(w http.ResponseWriter, r *http.Request) {
	likeUnlikeImage(w, r, false)
}

/**
[DELETE]

Deletes the given image.

Query parameters:
	- id: the image ID.

Returns
	- 200: Image deleted successfully.
	- 400: Image ID was not passed in.
	- 403: User does not have permission to delete image.
*/
func deleteImage(w http.ResponseWriter, r *http.Request) {

}

/**
[DELETE]
Deletes all images posted by this user.

Query parameters:
	- id: the image ID.

Returns
	- 200: Images deleted successfully. Will return the number of images deleted.
	- 403: User does not have permission to delete image.
*/
func deleteAllImages(w http.ResponseWriter, r *http.Request) {

}

/**
[GET]
Gets image by ID, and if user is authorized to see it.

Accepted query parameters:
	- id: Image ID.

Returns: (image/*)
		- 200 OK: With the provided image.
		- 400: If id is not present, or an invalid ID is passed in.
		- 404: If image is not found, or user is not authorised to view this image. (there is no difference).
*/
func getImage(w http.ResponseWriter, r *http.Request) {
	const parseErrorMessage = "Could not parse id parameter"

	imgIdArr, ok := r.URL.Query()["id"]

	if !ok || len(imgIdArr) < 1 || len(imgIdArr[0]) < 1 {
		http.Error(w, parseErrorMessage, http.StatusBadRequest)
		return
	}

	imgId := imgIdArr[0]

	hex, hexErr := primitive.ObjectIDFromHex(imgId)

	if hexErr != nil {
		http.Error(w, parseErrorMessage, http.StatusBadRequest)
		return
	}

	loggedInUser := getUserIDFromTokenNotStrictValidation(r)
	visibilityFilters := buildVisibilityFilters(loggedInUser)

	filter := bson.D{{"$and",
		[]interface{}{
			bson.D{{"$or", visibilityFilters}},
			bson.D{{"_id", hex},
			}}}}

	channel := make(chan imageDatabaseResponse)

	go getOneImage(filter, nil, channel)

	res := <-channel

	if res.err != nil {
		sendInternalServerError(w)
		return
	}

	if len(res.images) == 0 {
		http.Error(w, "Image not found", http.StatusNotFound)
		return
	}

	file, err := ioutil.TempFile("", imgId)

	if err != nil {
		log.Println(err)
		sendInternalServerError(w)
		return
	}

	defer os.Remove(file.Name())

	_, dlErr := s3Downloader.Download(file, &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(imgId),
	})

	if dlErr != nil {
		if strings.Contains(dlErr.Error(), "NoSuchKey") {
			http.Error(w, "Image not found", http.StatusNotFound)
		} else {
			log.Println(dlErr)
			sendInternalServerError(w)
		}
		return
	}

	fileContent, fileErr := ioutil.ReadFile(file.Name())

	if fileErr != nil {
		log.Println(fileErr)
		sendInternalServerError(w)
		return
	}

	w.Header().Add("Content-Type", http.DetectContentType(fileContent))
	w.WriteHeader(200)
	w.Write(fileContent)
}

/**
	[GET]

	Gets the metadata (not the actual image files) of the images in the database based on the queries passed in, in chronologically descending order.

	Accepted query parameters:
		- before: UNIX time stamp representing the latest image that can be uploaded.
		- after: UNIX time stamp repesenting the earliest image that should be fetched.
		- limit: integer. the limit on the number of images to fetch. Default 10 if not specified.
		- user: comma-separated string. Gets images from particular user(s).

	Returns: (application/json)
		- 200: With list of images that match search criteria.
		- 500: Internal server error.
*/
func getImagesMetadata(w http.ResponseWriter, r *http.Request) {
	filter, limit := buildImageQuery(r)

	opts := &options.FindOptions{
		Limit: &limit,
		Sort:  bson.D{{"uploadDateTime", -1}},
	}

	channel := make(chan imageDatabaseResponse)

	go getImagesMetadataFromDatabase(*filter, opts, channel)

	res := <-channel

	if res.err != nil {
		sendInternalServerError(w)
		return
	}

	appendAuthorsToImages(res.images, *res.userIDMap)

	marshalled, _ := json.Marshal(res.images)

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(200)
	_, _ = w.Write(marshalled)
}

/**
Initializes the AWS session, S3 Client S3 Uploader and S3 Downloader.
*/
func initAWS() {
	awsSession = session.Must(session.NewSessionWithOptions(session.Options{SharedConfigState: session.SharedConfigEnable}))
	s3Uploader = s3manager.NewUploader(awsSession)
	s3Downloader = s3manager.NewDownloader(awsSession)
}

func serveImageRoutes(r *mux.Router) {
	initAWS()

	r.HandleFunc("/getImage", getImage).Methods("GET")
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
