package images

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
	"strconv"
	"strings"
	"time"
)

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

