package model

type User struct {
	Name string		`json:"name,omitempty" bson:"name,omitempty"`
	Email string	`json:"emailAddr,omitempty" bson:"email,omitempty"`
	Password string	`json:"pwd,omitempty" bson:"password,omitempty"`
}
