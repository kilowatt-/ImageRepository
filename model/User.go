package model

type User struct {
	Name string		`json:"name,omitempty" bson:"name,omitempty"`
	Email string	`json:"emailAddr,omitempty" bson:"email,omitempty"`
	Password []byte	`json:"pwd,omitempty" bson:"password,omitempty"`
}
