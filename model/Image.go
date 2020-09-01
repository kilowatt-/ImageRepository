package model

// TODO: Add enums here
type Image struct {
	ID string `json:"_id,omitempty" bson:"_id,omitempty"`
	AuthorID string `json:"authorid,omitempty" bson:"authorid,omitempty"`
	AccessLevel string	`json:"accessLevel,omitempty" bson:"accessLevel,omitEmpty"`
	AccessListType string	`json:"accessListType,omitempty" bson:"accessListType,omitempty"`
	AccessListIDs []string	`json:"accessListIDs,omitempty" bson:"accessListIDs,omitempty"`
	Likes int	`json:"likes,omitempty" bson:"likes,omitempty"`
}