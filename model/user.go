package model

type User struct {
	ID      int   `json:"id"`
	WordIDs []int `json:"wordIDs"`
}