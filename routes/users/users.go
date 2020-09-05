package users

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/kilowatt-/ImageRepository/database"
	"github.com/kilowatt-/ImageRepository/model"
	"github.com/kilowatt-/ImageRepository/routes"
	"github.com/kilowatt-/ImageRepository/routes/middleware"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	"regexp"
	"strings"
)

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
			routes.SendInternalServerError(w)
		}
		return
	}

	projection := bson.D{{"userHandle", 1}, {"name", 1}}

	channel := make(chan []FindUserResponse)

	go GetUsersFromDatabase(query, projection, channel, opts)

	res := <-channel

	if len(res) == 1 && res[0].Err != nil {
		log.Println(res[0].Err)
		routes.SendInternalServerError(w)
		return
	}

	users := []model.User{}

	for _, k := range res {
		users = append(users, k.User)
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
			routes.SendInternalServerError(w)
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

		channel := make(chan FindUserResponse)
		go GetUserWithLogin(email, password, channel)

		res := <-channel

		if res.Err != nil {
			if res.Err.Error() == UserNotFound || res.Err.Error() == PasswordNotMatching {
				http.Error(w, "Provided email/password do not match", http.StatusNotFound)
			} else {
				routes.SendInternalServerError(w)
			}
		} else {
			token, expiry, jwtErr := middleware.CreateLoginToken(res.User)

			if jwtErr != nil {
				log.Println(jwtErr)
				routes.SendInternalServerError(w)
			} else {
				res.User.Password = nil
				jsonResponse, jsonErr := json.Marshal(res.User)

				if jsonErr != nil {
					log.Println(jsonErr)
					routes.SendInternalServerError(w)
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

func ServeUserRoutes(r *mux.Router) {
	r.HandleFunc("/signup", handleSignUp).Methods("POST")
	r.HandleFunc("/login", handleLogin).Methods("POST")
	r.HandleFunc("/getUsers", getUsers).Methods("GET")
}
