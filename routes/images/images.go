package images

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/kilowatt-/ImageRepository/database"
	"github.com/kilowatt-/ImageRepository/model"
	"github.com/kilowatt-/ImageRepository/routes/common"
	"github.com/kilowatt-/ImageRepository/routes/middleware"
	"github.com/kilowatt-/ImageRepository/routes/users"
	"github.com/kilowatt-/ImageRepository/util"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
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
var s3Instance *s3.S3 = nil
var s3Uploader *s3manager.Uploader = nil
var s3Downloader *s3manager.Downloader = nil

var mimeSet = map[string]bool{
	"image/bmp":  true,
	"image/gif":  true,
	"image/jpeg": true,
	"image/png":  true,
	"image/webp": true,
}

func getHexIdArray(acl *acl) (*[]primitive.ObjectID, error) {

	allowSet := make(map[primitive.ObjectID]bool)
	rejectSet := make(map[primitive.ObjectID]bool)

	for _, k := range acl.Add {
		idHex, hexErr := primitive.ObjectIDFromHex(k)

		if hexErr != nil {
			return nil, errors.New( "invalid user id: "+k)
		}

		allowSet[idHex] = true
	}

	for _, k := range acl.Remove {
		idHex, hexErr := primitive.ObjectIDFromHex(k)

		if hexErr != nil {
			return nil, errors.New( "invalid user id: "+k)
		}

		if _, exists := allowSet[idHex]; exists {
			return nil, errors.New("user id present in both add and delete sets: " + k)
		}

		rejectSet[idHex] = true
	}

	idArray := make([]primitive.ObjectID, 0, len(allowSet) + len(rejectSet))

	for k, _ := range allowSet {
		idArray = append(idArray, k)
	}

	return &idArray, nil
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

/**
Gets user ID from the token.

This is required for endpoints which might require a user ID, but are not checked by the middleware because they
are not strictly necessary. JWT Middleware is only run on functions where the user MUST be authenticated.
*/
func getUserIDFromTokenNotStrictValidation(r *http.Request) string {
	cookie, err := r.Cookie("token")
	if err == nil {
		token := cookie.Value
		valid, vErr := middleware.VerifyJWT(token)

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

func likeUnlikeImage(w http.ResponseWriter, r *http.Request, isLike bool) {
	uid := getUserIDFromToken(r)

	hex, err := getHexImageIDFromRequest(r)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	channel := make(chan *database.UpdateResponse)

	if isLike {
		go like(uid, *hex, channel)
	} else {
		go unlike(uid, *hex, channel)
	}

	res := <-channel

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

	w.WriteHeader(http.StatusOK)
}

/**
	[POST] form/multipart

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
		common.SendInternalServerError(w)
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

	log.Println(authorID)

	var accessListIDs []string

	jsonParseErr := json.Unmarshal([]byte(accessListIDsString), &accessListIDs)

	if jsonParseErr != nil {
		log.Println("passed empty access List IDs")
		accessListIDs = []string{}
	}

	if accessLevel != "public" && accessLevel != "private" {
		log.Println("invalid access level passed in; defaulting to public")
		accessLevel = "public"
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
		common.SendInternalServerError(w)
		return
	}
	id := insertResponse.ID

	_, uploadErr := s3Uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(id),
		Body:   buf,
	})

	if uploadErr != nil {
		common.SendInternalServerError(w)
		return
	}
	insertResponse.Err = nil

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	jsonResponse, _ := json.Marshal(insertResponse)
	_, _ = w.Write(jsonResponse)
}

func getHexImageIDFromRequest(r *http.Request) (*primitive.ObjectID, error) {
	image := &model.Image{}

	err := json.NewDecoder(r.Body).Decode(&image)

	if err != nil {
		return nil, err
	}

	if image.ID == "" {
		return nil, errors.New("image id not passed in")
	}

	hex, hexErr := primitive.ObjectIDFromHex(image.ID)

	if hexErr != nil {
		return nil, errors.New(invalidImageId)
	}

	return &hex, nil
}

func getHexIDFromString(str string) (*primitive.ObjectID, error) {
	hex, hexErr := primitive.ObjectIDFromHex(str)

	if hexErr != nil {
		return nil, errors.New(invalidImageId)
	}

	return &hex, nil
}


type acl struct {
	ID		string	`json:"_id,omitempty", bson:"_id,omitempty"`
	Add    []string `json:"add,omitempty" bson:"add,omitempty"`
	Remove []string `json:"remove,omitempty" bson:"remove,omitempty"`
}

/**
	[PATCH]

	Adds the selected user IDs to the image's access control list (ACL)

	JSON body parameters:
		- _id: the image ID.
		- add: an array of strings: the user IDs to add.
		- remove: an array of strings: user IDs to remove from database.

	Returns:
		- 200 OK: All users were added/removed to the ACL.
		- 204 No Content: ACL was not modified.
		- 400: Invalid image ID was sent, at least one invalid user ID was passed in, same user ID was present in both add and delete lists, or both add and remove lists are empty.
		- 404: At least one user in the add and remove lists does not exist in the database, or image not found.
		- 500: Internal server error
 */
func editImageACL(w http.ResponseWriter, r *http.Request) {
	uid := getUserIDFromToken(r)

	acl := &acl{}

	err := json.NewDecoder(r.Body).Decode(&acl)

	acl.Add = util.RemoveDuplicatesFromStringArray(acl.Add)
	acl.Remove = util.RemoveDuplicatesFromStringArray(acl.Remove)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	hex, err := getHexIDFromString(acl.ID)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(acl.Add) == 0 && len(acl.Remove) == 0 {
		 http.Error(w, "no users passed to add/remove array", http.StatusBadRequest)
		 return
	}

	idArray, idErr := getHexIdArray(acl)

	if idErr != nil {
		http.Error(w, idErr.Error(), http.StatusBadRequest)
		return
	}

	usersFilter := bson.D{{"_id", bson.D{{"$in", idArray}}}}
	usersChannel := make(chan []users.FindUserResponse)

	go users.GetUsersFromDatabase(usersFilter, nil, usersChannel)

	res := <- usersChannel

	if len(res) != len(*idArray) {
		http.Error(w, "not all users in add/remove list found", http.StatusNotFound)
		return
	}

	addChan := make(chan *database.UpdateResponse)
	go updateACLAdd(*hex, uid, acl.Add, addChan)
	responseAdd := <-addChan

	if responseAdd.Matched == 0 {
		http.Error(w, "image not found", http.StatusNotFound)
		return
	}

	if responseAdd.Err != nil {
		log.Println(responseAdd.Err)
		common.SendInternalServerError(w)
		return
	}

	rmChan := make(chan *database.UpdateResponse)
	go updateACLRemove(*hex, uid, acl.Remove, rmChan)
	responseRemove := <-rmChan

	if responseRemove.Err != nil {
		log.Println(responseRemove.Err)
		common.SendInternalServerError(w)
		return
	}

	if responseAdd.Modified == 0 && responseRemove.Modified == 0{
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.WriteHeader(http.StatusOK)
}



/**
[PUT]

Adds this user to the image's like list.

JSON body parameters:
	- _id: the image ID.

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

Removes this user from the image's like list.

Query parameters:
	- _id: the image ID.

Returns:
	- 200: Image liked successfully.
	- 400: id not passed in.
	- 404: Image not found, or user not authorised to view image. (no difference.)
	- 409: User already unliked image. No-op.
*/
func unlikeImage(w http.ResponseWriter, r *http.Request) {
	likeUnlikeImage(w, r, false)
}

/**
[DELETE]

Deletes the given image.

JSON body parameters:
	- id: the image ID.

Returns
	- 200: Image deleted successfully.
	- 400: Image ID was not passed in.
	- 404: Image not found or user does not have permission to delete image. (no difference)
	- 500: Internal server error
*/
func deleteImage(w http.ResponseWriter, r *http.Request) {
	uid := getUserIDFromToken(r)

	hex, err := getHexImageIDFromRequest(r)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	channel := make(chan *database.DeleteResponse)

	go deleteImageFromDatabase(uid, *hex, channel)

	res := <-channel

	if res.Err != nil {
		common.SendInternalServerError(w)
		return
	}

	if res.NumberDeleted == 0 {
		http.Error(w, "image not found", http.StatusNotFound)
		return
	}

	if _, err := s3Instance.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(bucketName),
		Key: aws.String(hex.Hex()),
	}); err != nil {
		common.SendInternalServerError(w)
		return
	}

	w.WriteHeader(http.StatusOK)
}

/**
[GET]
Gets image by ID, if it is visible to the user.

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
		common.SendInternalServerError(w)
		return
	}

	if len(res.images) == 0 {
		http.Error(w, "Image not found", http.StatusNotFound)
		return
	}

	file, err := ioutil.TempFile("", imgId)

	if err != nil {
		log.Println(err)
		common.SendInternalServerError(w)
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
			common.SendInternalServerError(w)
		}
		return
	}

	fileContent, fileErr := ioutil.ReadFile(file.Name())

	if fileErr != nil {
		log.Println(fileErr)
		common.SendInternalServerError(w)
		return
	}

	w.Header().Add("Content-Type", http.DetectContentType(fileContent))
	w.WriteHeader(http.StatusOK)
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
		- id: comma-separated string. Image ID(s). All other parameters are ignored if this is possed in.

	Returns: (application/json)
		- 200: With list of images that match search criteria.
		- 400: If an invalid hex ID was passed in.
		- 500: Internal server error.
*/
func getImagesMetadata(w http.ResponseWriter, r *http.Request) {
	if filter, limit, hexErr := buildImageQuery(r); hexErr != nil {
		http.Error(w, hexErr.Error(), http.StatusBadRequest)
	} else {
		opts := &options.FindOptions{
			Limit: &limit,
			Sort:  bson.D{{"uploadDateTime", -1}},
		}

		channel := make(chan imageDatabaseResponse)

		go getImagesMetadataFromDatabase(*filter, opts, channel)

		res := <-channel

		if res.err != nil {
			common.SendInternalServerError(w)
			return
		}

		appendAuthorsToImages(res.images, *res.userIDMap)

		marshalled, _ := json.Marshal(res.images)

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(marshalled)
	}
}

/**
Initializes the AWS session, S3 Client S3 Uploader and S3 Downloader.
*/
func initAWS() {
	awsSession = session.Must(session.NewSessionWithOptions(session.Options{SharedConfigState: session.SharedConfigEnable}))
	s3Instance = s3.New(awsSession)
	s3Uploader = s3manager.NewUploader(awsSession)
	s3Downloader = s3manager.NewDownloader(awsSession)
}

func ServeImageRoutes(r *mux.Router) {
	initAWS()

	r.HandleFunc("/getImage", getImage).Methods("GET")
	r.HandleFunc("/getImagesMetadata", getImagesMetadata).Methods("GET")

	s := r.Methods("DELETE", "PATCH", "POST").Subrouter()

	s.Use(middleware.JWTMiddleware)

	s.HandleFunc("/addImage", addNewImage).Methods("POST")
	s.HandleFunc("/deleteImage", deleteImage).Methods("DELETE")
	s.HandleFunc("/editImageACL", editImageACL).Methods("PATCH")
	s.HandleFunc("/likeImage", likeImage).Methods("PATCH")
	s.HandleFunc("/unlikeImage", unlikeImage).Methods("DELETE")
}



