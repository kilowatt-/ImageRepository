package routes

import (
	"encoding/json"
	"errors"
	"github.com/gorilla/mux"
	"github.com/kilowatt-/ImageRepository/database"
	"github.com/kilowatt-/ImageRepository/model"
	routes "github.com/kilowatt-/ImageRepository/routes/middleware"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
	"log"
	"math"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

const UserNotFound = "user not found"
const PasswordNotMatching = "password does not match"
const InvalidHex = "invalid hex"

type findUserResponse struct {
	user model.User
	err  error
}

func createUser(user model.User, channel chan *database.InsertResponse) {
	res := database.InsertOne("users", user, nil)

	channel <- res
}

/**
Builds a user query based on the inputs to getUsers API endpoint.

Returns the built query, the projection direction, the options, and an error (if present).
*/
func buildUserQueryAndOptions(r *http.Request) (*bson.D, *options.FindOptions, error) {
	ids := []primitive.ObjectID{}
	names := []string{}
	handles := []string{}
	var limit int64 = 100

	namesExact := false
	handlesExact := false
	nameAsOrder := false
	direction := -1

	lt := ""
	gt := ""

	if idQuery, ok := r.URL.Query()["id"]; ok && len(idQuery) > 0 && len(idQuery[0]) > 0 {
		idStrs := strings.Split(idQuery[0], ",")

		for _, k := range idStrs {
			hex, err := primitive.ObjectIDFromHex(k)

			if err != nil {
				return nil, nil, errors.New(InvalidHex)
			}

			ids = append(ids, hex)
		}
	}

	if limitQuery, ok := r.URL.Query()["limit"]; ok && len(limitQuery) > 0 && len(limitQuery[0]) > 0 {
		if parseLimit, err := strconv.ParseInt(limitQuery[0], 10, 64); err == nil {
			limit = int64(math.Min(float64(parseLimit), float64(limit)))
		}
	}

	if orderByQuery, ok := r.URL.Query()["orderBy"]; ok && len(orderByQuery) > 0 && len(orderByQuery[0]) > 0 {
		nameAsOrder = orderByQuery[0] == "name"
	}

	if ascQuery, ok := r.URL.Query()["ascending"]; ok && len(ascQuery) > 0 && len(ascQuery[0]) > 0 && (ascQuery[0] == "Y" || ascQuery[0] == "y") {
		direction = 1
	}

	opts := &options.FindOptions{Limit: &limit}

	if nameAsOrder {
		opts.SetSort(bson.D{{"name", direction}, {"userHandle", direction}})
	} else {
		opts.SetSort(bson.D{{"userHandle", direction}, {"name", direction}})
	}

	if len(ids) > 0 {
		opts.SetLimit(int64(len(ids)))
		query := &bson.D{{"_id", bson.D{{"$in", ids}}}}
		return query, opts, nil
	}

	subQueries := []bson.D{{}}

	if ltQuery, ok := r.URL.Query()["lt"]; ok && len(ltQuery) > 0 && len(ltQuery[0]) > 0 {
		lt = ltQuery[0]
	}

	if gtQuery, ok := r.URL.Query()["gt"]; ok && len(gtQuery) > 0 && len(gtQuery[0]) > 0 {
		gt = gtQuery[0]
	}

	if nameExactQuery, ok := r.URL.Query()["nameExact"]; ok && len(nameExactQuery) > 0 && len(nameExactQuery[0]) > 0 {
		namesExact = nameExactQuery[0] == "y" || nameExactQuery[0] == "Y"
	}

	if userHandleExactQuery, ok := r.URL.Query()["userHandleExact"]; ok && len(userHandleExactQuery) > 0 && len(userHandleExactQuery[0]) > 0 {
		handlesExact = userHandleExactQuery[0] == "y" || userHandleExactQuery[0] == "Y"
	}

	if namesQuery, ok := r.URL.Query()["name"]; ok && len(namesQuery) > 0 && len(namesQuery[0]) > 0 {
		names = strings.Split(namesQuery[0], ",")

		baseQuery := bson.D{{"name", bson.D{{"$in", names}}}}

		if !namesExact {
			namesRegEx := make([]primitive.Regex, len(names))

			for i, k := range names {
				namesRegEx[i] = primitive.Regex{
					Pattern: k,
					Options: "i",
				}
			}

			baseQuery = bson.D{{"name", bson.D{{"$in", namesRegEx}}}}
		}

		if nameAsOrder {
			q := []bson.D{baseQuery}

			if lt != "" {
				q = append(q, bson.D{{"name", bson.D{{"$lt", lt}}}})
			}

			if gt != "" {
				q = append(q, bson.D{{"name", bson.D{{"$gt", gt}}}})
			}

			subQueries = append(subQueries, bson.D{{"$and", q}})
		} else {
			subQueries = append(subQueries, baseQuery)
		}
	}

	if handlesQuery, ok := r.URL.Query()["userHandle"]; ok && len(handlesQuery) > 0 && len(handlesQuery[0]) > 0 {
		handles = strings.Split(handlesQuery[0], ",")
		baseQuery := bson.D{{"userHandle", bson.D{{"$in", handles}}}}

		if !handlesExact {
			handlesRegEx := make([]primitive.Regex, len(handles))

			for i, k := range handles {
				handlesRegEx[i] = primitive.Regex{
					Pattern: k,
					Options: "i",
				}
			}

			baseQuery = bson.D{{"userHandle", bson.D{{"$in", handlesRegEx}}}}
		}

		if !nameAsOrder {
			q := []bson.D{baseQuery}

			if lt != "" {
				q = append(q, bson.D{{"userHandle", bson.D{{"$lt", lt}}}})
			}

			if gt != "" {
				q = append(q, bson.D{{"userHandle", bson.D{{"$gt", gt}}}})
			}

			subQueries = append(subQueries, bson.D{{"$and", q}})
		} else {
			subQueries = append(subQueries, baseQuery)
		}

	}

	query := &bson.D{{"$and", subQueries}}

	return query, opts, nil
}

/**
Gets users from the database according to the given filter and projection.
*/
func getUsersFromDatabase(filter interface{}, projection interface{}, channel chan []findUserResponse, optionsList ...*options.FindOptions) {
	opts := &options.FindOptions{}

	if projection != nil {
		opts.SetProjection(projection)
	}

	if len(optionsList) > 0 {
		for _, o := range optionsList {
			opts = options.MergeFindOptions(opts, o)
		}
	}

	var userArray = []findUserResponse{}

	res := database.Find("users", filter, opts)

	if res.Err != nil {
		userArray = append(userArray, findUserResponse{model.User{}, res.Err})
	} else {
		for i := 0; i < len(res.Result); i++ {
			var user model.User
			bsonBytes, _ := bson.Marshal(res.Result[i])
			_ = bson.Unmarshal(bsonBytes, &user)

			userArray = append(userArray, findUserResponse{user, nil})
		}
	}

	channel <- userArray
}

/**
Gets user from database based on email and password.

If user is not found, an error will be returned.
*/
func getUserWithLogin(email string, password string, channel chan findUserResponse) {
	filter := bson.D{{"email", email}}

	blankUser := model.User{
		Name:     "",
		Email:    "",
		Password: nil,
	}

	response := database.FindOne("users", filter, nil)

	if response.Err != nil {
		channel <- findUserResponse{user: blankUser, err: response.Err}
	} else if len(response.Result) == 0 {
		channel <- findUserResponse{user: blankUser, err: errors.New(UserNotFound)}
	} else {
		var user model.User

		bsonBytes, _ := bson.Marshal(response.Result)

		_ = bson.Unmarshal(bsonBytes, &user)

		pwdErr := bcrypt.CompareHashAndPassword(user.Password, []byte(password))

		if pwdErr != nil {
			channel <- findUserResponse{user: blankUser, err: errors.New(PasswordNotMatching)}
		} else {
			channel <- findUserResponse{user, nil}
		}
	}
}

func verifyEmail(email string) bool {
	matched, err := regexp.MatchString("(?:[a-z0-9!#$%&'*+/=?^_`{|}~-]+(?:\\.[a-z0-9!#$%&'*+/=?^_`{|}~-]+)*|\"(?:[\\x01-\\x08\\x0b\\x0c\\x0e-\\x1f\\x21\\x23-\\x5b\\x5d-\\x7f]|\\\\[\\x01-\\x09\\x0b\\x0c\\x0e-\\x7f])*\")@(?:(?:[a-z0-9](?:[a-z0-9-]*[a-z0-9])?\\.)+[a-z0-9](?:[a-z0-9-]*[a-z0-9])?|\\[(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?|[a-z0-9-]*[a-z0-9]:(?:[\\x01-\\x08\\x0b\\x0c\\x0e-\\x1f\\x21-\\x5a\\x53-\\x7f]|\\\\[\\x01-\\x09\\x0b\\x0c\\x0e-\\x7f])+)\\])", email)
	return err == nil && matched
}

func verifyUserHandle(userHandle string) bool {
	matched, err := regexp.MatchString("^[a-zA-Z0-9]+$", userHandle)
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

	for i := 0; i < pwdLength; i++ {
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

/**
[GET]
Gets users that match the query. An empty query will return the first 100 users.

Accepted query parameters:
	- id: {comma separated hex strings} A list of user IDs. Will ignore all queries except ascending and orderBy if this is present.
	- name: {comma separated hex strings} A list of names. Non-exact, case-insensitive match by default.
	- nameExact: {Y/y} A flag that sets whether the database query should match the names exactly. Set this flag to Y/y only if you want to get exact matches
	- userHandle: {comma separated strings} A list of user handles. Non-exact, case-insensitive match by default.
	- userHandleExact: {Y/y} A flag that sets whether the query should match the userHandles exactly or not. Set this flag to Y/y only if you want to get exact matches
	- limit: {int}: Limit on the number of results returned. Default (and max) of 100.
	- ascending: {Y/y}: Whether to return the results in ascending order or not.
	- orderBy: {userHandle/name}: Whether to order the name by userHandle or name. Default is userHandle. The other value is treated as the tiebreaker.
	- lt: {string}: An *exact* string that represents that values less than this string should be returned. Corresponds to the ordering key (userHandle or name)
	- gt: {string}: An *exact* string that represents values more than this string should be returned.  Corresponds to the ordering key (userHandle or name)
Returns:
	- 200: List of matching users that match the query parameters.
	- 400: If at least one of the IDs passed in is invalid.
	- 500: Internal server error.
*/
func getUsers(w http.ResponseWriter, r *http.Request) {
	query, opts, err := buildUserQueryAndOptions(r)

	if err != nil {
		if err.Error() == InvalidHex {
			http.Error(w, "invalid ID passeed in", http.StatusBadRequest)
		} else {
			sendInternalServerError(w)
		}
		return
	}

	projection := bson.D{{"userHandle", 1}, {"name", 1}}

	channel := make(chan []findUserResponse)

	go getUsersFromDatabase(query, projection, channel, opts)

	res := <-channel

	if len(res) == 1 && res[0].err != nil {
		log.Println(res[0].err)
		sendInternalServerError(w)
		return
	}

	users := []model.User{}

	for _, k := range res {
		users = append(users, k.user)
	}

	jsonResponse, _ := json.Marshal(users)
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(jsonResponse)
}

/**
[POST]
Handles sign up. Takes in a name, email, userHandle and password, verifies the inputs, and creates the user.

Returns:
- 200: User was created. Returns ID.
- 400: If any complexity requirement was not met, an invalid email was sent, or an invalid form was sent.
- 409: Conflict, if the user was already registered with the email or userHandle.
- 500: If there is an error on the server (Database error, etc).
*/
func handleSignUp(w http.ResponseWriter, r *http.Request) {
	if parseFormErr := r.ParseForm(); parseFormErr != nil {
		http.Error(w, "Sent invalid form", 400)
	}

	name := r.FormValue("name")
	userHandle := r.FormValue("userHandle")
	email := r.FormValue("email")
	password := r.FormValue("password")

	if !verifyUserHandle(userHandle) {
		http.Error(w, "Invalid userHandle", http.StatusBadRequest)
	}

	if !verifyEmail(email) {
		http.Error(w, "Invalid email", http.StatusBadRequest)
		return
	}

	if !verifyPassword(password) {
		http.Error(w, "Password does not meet complexity requirements", http.StatusBadRequest)
		return
	}

	hashed, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	urChannel := make(chan *database.InsertResponse)
	go createUser(
		model.User{Name: name, UserHandle: userHandle, Email: email, Password: hashed},
		urChannel,
	)
	createdUser := <-urChannel

	if createdUser.Err != nil {
		log.Println(createdUser.Err)

		if strings.Contains(createdUser.Err.Error(), "E11000") {
			if strings.Contains(createdUser.Err.Error(), "index: userHandle_1") {
				http.Error(w, "Userhandle "+userHandle+" already registered", http.StatusConflict)
			} else {
				http.Error(w, "Email "+email+" already registered", http.StatusConflict)
			}
		} else {
			sendInternalServerError(w)
		}

	} else {
		log.Println("Created user with ID " + createdUser.ID)
		w.WriteHeader(http.StatusOK)
		_, wError := w.Write([]byte("Created user with ID " + createdUser.ID))

		if wError != nil {
			log.Println("Error while writing: " + wError.Error())
		}
	}

}

/**
[POST]
Handles a login request.

Returns:
- 200: OK, if the username and password match. Will return the userinfo, and set userinfo and JWT cookies
- 404: If the user was not found, or the username and password don't match. (there is no difference here.)
- 500: If there is an internal server error.
*/
func handleLogin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	parseFormErr := r.ParseForm()

	if parseFormErr != nil {
		http.Error(w, "Sent invalid form", http.StatusBadRequest)
	} else {
		email := r.FormValue("email")
		password := r.FormValue("password")

		channel := make(chan findUserResponse)
		go getUserWithLogin(email, password, channel)

		res := <-channel

		if res.err != nil {
			if res.err.Error() == UserNotFound || res.err.Error() == PasswordNotMatching {
				http.Error(w, "Provided email/password do not match", http.StatusNotFound)
			} else {
				sendInternalServerError(w)
			}
		} else {
			token, expiry, jwtErr := routes.CreateLoginToken(res.user)

			if jwtErr != nil {
				log.Println(jwtErr)
				sendInternalServerError(w)
			} else {
				res.user.Password = nil
				jsonResponse, jsonErr := json.Marshal(res.user)

				if jsonErr != nil {
					log.Println(jsonErr)
					sendInternalServerError(w)
				} else {

					jsonEncodedCookie := strings.ReplaceAll(string(jsonResponse), "\"", "'") // Have to do this to Set-Cookie in psuedo-JSON format.

					http.SetCookie(w, &http.Cookie{
						Name:       "token",
						Value:      token,
						Path:       "/",
						Expires:    expiry,
						RawExpires: expiry.String(),
						Secure:     false,
						HttpOnly:   true,
						SameSite:   0,
					})

					http.SetCookie(w, &http.Cookie{
						Name:       "userinfo",
						Value:      jsonEncodedCookie,
						Path:       "/",
						Expires:    expiry,
						RawExpires: expiry.String(),
						Secure:     false,
						HttpOnly:   false,
						SameSite:   0,
					})

					w.WriteHeader(200)
					w.Write(jsonResponse)
				}
			}
		}
	}
}

func serveUserRoutes(r *mux.Router) {
	r.HandleFunc("/signup", handleSignUp).Methods("POST")
	r.HandleFunc("/login", handleLogin).Methods("POST")
	r.HandleFunc("/getUsers", getUsers).Methods("GET")
}
