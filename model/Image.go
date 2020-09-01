package model

// TODO: Add enums here
type Image struct {
	ID string `json:"_id,omitempty" bson:"_id,omitempty"`
	AuthorID string `json:"authorid,omitempty" bson:"authorid,omitempty"`
	Width int	`json:"width,omitempty" bson:"width,omitempty"`
	Height int	`json:"height,omitempty" bson:"height,omitempty"`
	AccessLevel string	`json:"accessLevel,omitempty" bson:"accessLevel,omitEmpty"`
	AccessListType string	`json:"accessListType,omitempty" bson:"accessListType,omitempty"`
	AccessListIDs []string	`json:"accessListIDs,omitempty" bson:"accessListIDs,omitempty"`
	Likes int	`json:"likes,omitempty" bson:"likes,omitempty"`
	Base64 string	`json:"base64,omitempty" bson:"base64,omitempty"`
	URL	string	`json:"url,omitempty" bson:"url,omitempty"`
}