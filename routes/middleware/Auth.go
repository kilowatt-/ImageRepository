package routes

import (
	"errors"
	"github.com/dgrijalva/jwt-go"
	"github.com/kilowatt-/ImageRepository/model"
	"net/http"
	"os"
	"strings"
	"time"
)

const JWTKeyNotFound = "jwt key not found"


/**
	Verifies if the given token is valid.
 */
func verifyJWT(token string, secretKey string) (bool, error) {
	t, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})

	return t.Valid, err
}

/**
	Middleware function.
 */
func JWTMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		extractedToken := strings.Split(token, "Bearer ")

		valid, _ := verifyJWT(strings.Join(extractedToken, ""), os.Getenv("JWT_KEY"))

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
	secretKey, keyExists := os.LookupEnv("JWT_KEY")

	if !keyExists {
		return "", time.Now(), errors.New(JWTKeyNotFound)
	}

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

