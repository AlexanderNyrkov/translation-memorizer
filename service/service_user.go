package service

import (
	"fmt"
	"github.com/anyrkov/translation-memorizer/common"
	"github.com/anyrkov/translation-memorizer/model"
	"gopkg.in/mgo.v2/bson"
)

const usersCollection = "users"

func CreateUser(id int, user *model.User) {
	dbConn := common.GetSession().Copy()
	defer dbConn.Close()

	err := dbConn.DB("").C(usersCollection).Find(bson.M{"id": id}).One(&user)

	if err != nil {
		user.ID = id
		if err := dbConn.DB("").C(usersCollection).Insert(user); err != nil {
			panic(fmt.Sprintf("I can't create this user - %d", id))
		}
	}
}

func AddWordToUser(id int, wordId int) {
	user := getUser(id)

	words := append(user.WordIDs, wordId)

	updateUserWordIDs(id, words)
}

func GetAllUserWords(id int) []int {
	return getUser(id).WordIDs
}

func IsUserHasWords(id int) bool {
	if len(getUser(id).WordIDs) > 0 {
		return true
	}

	return false
}

func deleteWordFromUser(id int, wordId int) {
	user := getUser(id)

	for i, word := range user.WordIDs {
		if word == wordId {
			user.WordIDs = append(user.WordIDs[:i], user.WordIDs[i+1:]...)
			break
		}
	}

	updateUserWordIDs(id, user.WordIDs)
}

func getUser(id int) model.User {
	var user model.User
	dbConn := common.GetSession().Copy()
	defer dbConn.Close()
	if err := dbConn.DB("").C(usersCollection).Find(bson.M{"id": id}).One(&user); err != nil {
		panic(fmt.Sprintf("The user - %d doesn't exist", id))
	}
	return user
}

func updateUserWordIDs(id int, wordIDs []int) {
	dbConn := common.GetSession().Copy()
	defer dbConn.Close()
	if err := dbConn.DB("").C(usersCollection).Update(bson.M{"id": id}, bson.M{"$set": bson.M{"wordids": wordIDs}}); err != nil {
		panic(fmt.Sprintf("I can't update a %d user words", id))
	}
}

