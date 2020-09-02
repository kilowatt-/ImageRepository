package routes

import (
	"errors"
	"github.com/dgrijalva/jwt-go"
	"github.com/kilowatt-/ImageRepository/database"
	"github.com/kilowatt-/ImageRepository/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
	"net/http"
	"os"
	"time"
)


/**
	Verifies if the given token is valid.
 */
func verifyJWT(token string, secretKey string) (bool, error) {
	t, err := jwt.ParseWithClaims(token, &jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})

	if t != nil  {
		id :=  (*(t.Claims.(*jwt.MapClaims)))["id"].(string)

		if id == "" {
			return false, errors.New("no id in token")
		}

		primitiveID, hexErr := primitive.ObjectIDFromHex(id)

		if hexErr != nil {
			return false, hexErr
		}

		user, dbErr := database.FindOne("users", bson.D{{"_id", primitiveID}})

		if dbErr != nil {
			return false, err
		}

		if len(user) == 0 {
			return false, errors.New("user not found")
		}

		return t.Valid, err
	}

	return false, errors.New("malformed token presented")
}

/**
	Middleware function for JWTs. Validates incoming JWTs and returns 401 if they are unauthorized.

	This is meant to be used with endpoints that REQUIRE a login; endpoints that might require a JWT for access to
	protected resources require another function to be called.
 */
func JWTMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, cErr := r.Cookie("token")

		if cErr != nil  {
			log.Println(cErr)
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		token := cookie.Value

		valid, _ := verifyJWT(token, os.Getenv("JWT_KEY"))

		if !valid {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
		} else {
			next.ServeHTTP(w, r)
		}
	})
}

/**
	Creates a login token that is a JWT. Expires 1 hour after creation.

	Returns the signed token, expiry date, and error.
*/
func CreateLoginToken(user model.User) (string, time.Time, error) {
	secretKey := os.Getenv("JWT_KEY")

	now := time.Now()
	expiry := now.Add(time.Hour * 1)

	claims := jwt.MapClaims{}

	claims["authorized"] = true
	claims["id"] = user.ID
	claims["email"] = user.Email
	claims["name"] = user.Name
	claims["loginTime"] = now.Unix()
	claims["exp"] = expiry.Unix()

	unsignedJwt := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	token, err := unsignedJwt.SignedString([]byte(secretKey))

	if err != nil {
		return "", time.Now(), err
	}

	return token, expiry, nil
}

