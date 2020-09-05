package images

import (
	"github.com/kilowatt-/ImageRepository/database"
	"github.com/kilowatt-/ImageRepository/model"
	"github.com/kilowatt-/ImageRepository/routes/users"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func deleteAllImagesFromDatabase(userid string, channel chan *database.DeleteResponse) {
	filter := bson.D{{"authorid", userid}}

	channel <- database.Delete("images", filter, nil)
}

func deleteImageFromDatabase(userid string, imageid primitive.ObjectID, channel chan *database.DeleteResponse) {
	filter := bson.D{{"$and", []bson.D{
		{{"authorid", userid}},
		{{"_id", imageid}},
	}}}

	channel <- database.DeleteOne("images", filter, nil)
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
	c := make(chan []users.FindUserResponse)
	go users.GetUsersFromDatabase(filter, projection, c)

	res := <-c

	if len(res) == 1 && res[0].Err != nil {
		return
	}

	userMap := make(map[string]model.User)

	for i := 0; i < len(res); i++ {
		cur := res[i].User
		userMap[cur.ID] = cur
	}

	for i := 0; i < len(images); i++ {
		cur := images[i]
		cur.SetAuthor(userMap[cur.AuthorID])
	}
}

