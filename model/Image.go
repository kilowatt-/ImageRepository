package model

// TODO: Add enums here
type Image struct {
	ID string
	AuthorID string
	Width int
	Height int
	FileType string
	URL string
	AccessLevel string
	AccessListType string
	AccessListIDs []string
	Likes int
}