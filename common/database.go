package common

import (
	"gopkg.in/mgo.v2"
)

func GetSession() *mgo.Session {
	s, err := mgo.Dial("mongodb://localhost")
	if err != nil {
		panic(err)
	}
	return s
}
