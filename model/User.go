package model

type User struct {
	ID string		`json:"_id,omitempty" bson:"_id,omitempty"`
	Name string		`json:"name,omitempty" bson:"name,omitempty"`
	Email string	`json:"emailAddr,omitempty" bson:"email,omitempty"`
	Password []byte	`json:"pwd,omitempty" bson:"password,omitempty"`
}
