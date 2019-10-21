package model

type Word struct {
	ID           int    `json:"id"`
	Original     string `json:"original"`
	Translation  string `json:"translation"`
	IsTranslated bool   `json:"isTranslated"`
}
