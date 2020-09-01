package routes

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/kilowatt-/ImageRepository/model"
	"net/http"
	"os"
	"strings"
	"time"
)


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
	Middleware function for JWTs. Validates incoming JWTs and returns 401 if they are unauthorized.

	This is meant to be used with endpoints that REQUIRE a login; endpoints that might require a JWT for access to
	protected resources require another function to be called.
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

