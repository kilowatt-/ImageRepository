package model

import "time"

// TODO: Add enums here
type Image struct {
	ID string `json:"_id,omitempty" bson:"_id,omitempty"`
	AuthorID string `json:"authorid,omitempty" bson:"authorid,omitempty"`
	AccessLevel string	`json:"accessLevel,omitempty" bson:"accessLevel,omitEmpty"`
	AccessListIDs []string	`json:"accessListIDs,omitempty" bson:"accessListIDs,omitempty"`
	Likes int	`json:"likes,omitempty" bson:"likes,omitempty"`
	Caption string `json:"caption,omitempty" bson:"caption,omitempty"`
	UploadDate time.Time `json:"uploadDateTime,omitempty" bson:"uploadDateTime,omitempty"`
}