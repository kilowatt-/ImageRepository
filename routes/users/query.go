package users

import (
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"math"
	"net/http"
	"strconv"
	"strings"
)

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

	opts := &options.FindOptions{Limit: &limit}

	if orderByQuery, ok := r.URL.Query()["orderBy"]; ok && len(orderByQuery) > 0 && len(orderByQuery[0]) > 0 {
		nameAsOrder = orderByQuery[0] == "name"
	}

	if ascQuery, ok := r.URL.Query()["ascending"]; ok && len(ascQuery) > 0 && len(ascQuery[0]) > 0 && (ascQuery[0] == "Y" || ascQuery[0] == "y") {
		direction = 1
	}

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

