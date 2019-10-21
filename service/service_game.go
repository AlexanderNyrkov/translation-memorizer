package service

import (
	"bytes"
	"fmt"
	"github.com/anyrkov/translation-memorizer/common"
	"github.com/anyrkov/translation-memorizer/model"
	"gopkg.in/mgo.v2/bson"
	"gopkg.in/telegram-bot-api.v4"
	"math/rand"
	"strings"
	"time"
)

const gameCollection = "games"

func StartGame(update tgbotapi.Update, message *tgbotapi.MessageConfig) {

	var word model.Word

	if HasWordWithoutTranslation(update.Message.From.ID, &word) {
		DeleteWord(update.Message.From.ID, word)
	}

	if !IsUserHasWords(update.Message.From.ID) {
		message.Text = "You don't have a words"
		return
	}

	game := getCurrentGame(update.Message.From.ID)

	if game.IsActive {
		message.Text = "Game already started"
		return
	}

	gameTypes := []tgbotapi.KeyboardButton{
		{Text: model.OriginalGame.String()},
		{Text: model.TranslateGame.String()},
		{Text: model.MixedGame.String()},
	}

	message.ReplyMarkup = &tgbotapi.ReplyKeyboardMarkup{
		Keyboard:        [][]tgbotapi.KeyboardButton{gameTypes},
		ResizeKeyboard:  true,
		OneTimeKeyboard: true,
		Selective:       false,
	}

	game.UserId = update.Message.From.ID
	game.IsActive = true

	dbConn := common.GetSession().Copy()
	defer dbConn.Close()

	if err := dbConn.DB("").C(gameCollection).Insert(game); err != nil {
		panic(fmt.Sprintf("I can't create a game for this user : %d", game.UserId))
	}

	message.Text = "Choose a game"
}

func StopGame(update tgbotapi.Update, message *tgbotapi.MessageConfig) {
	game := getCurrentGame(update.Message.From.ID)

	if !game.IsActive {
		message.Text = "The game does't started yet"
		return
	}

	message.Text = fmt.Sprintf("Game over! *Your result is: %d* , *attempts: %d*", game.CountRightAnswers, game.CountAttempts)
	message.ParseMode = "markdown"

	dbConn := common.GetSession().Copy()
	defer dbConn.Close()

	if err := dbConn.DB("").C(gameCollection).Remove(game); err != nil {
		panic(fmt.Sprintf("I can't delete a game for this user : %d", game.UserId))
	}
}

func PlayGame(update tgbotapi.Update, message *tgbotapi.MessageConfig) {
	var reply bytes.Buffer
	game := getCurrentGame(update.Message.From.ID)
	dbConn := common.GetSession().Copy()
	defer dbConn.Close()

	message.ReplyMarkup = tgbotapi.ReplyKeyboardRemove{RemoveKeyboard:true}

	if game.Type == "" {
		game.Type = model.GameType(update.Message.Text)
	}

	if len(game.Answer) > 0 {
		if strings.EqualFold(game.Answer, update.Message.Text) {
			game = clearQuestionsAndAnswers(game)
			if err := dbConn.DB("").C(gameCollection).Update(bson.M{"userid": game.UserId}, bson.M{"$set": bson.M{"countrightanswers": game.CountRightAnswers+1}}); err != nil {
				panic(fmt.Sprintf("I can't add a right answers for this user : %d", game.UserId))
			}
			reply.WriteString("Very good !\n")
		} else {
			reply.WriteString(fmt.Sprintf("Uhhh wrong!, right answer is *%s* !\n", game.Answer))

			game = clearQuestionsAndAnswers(game)
		}

		if err := dbConn.DB("").C(gameCollection).Update(bson.M{"userid": game.UserId}, bson.M{"$set": bson.M{"countattempts": game.CountAttempts+1}}); err != nil {
			panic(fmt.Sprintf("I can't add an attempts for this user : %d", game.UserId))
		}
	}

	if len(game.Question) == 0 {
		wordIDs := GetAllUserWords(update.Message.From.ID)

		rand.Seed(time.Now().Unix())
		word := getWordById(wordIDs[rand.Intn(len(wordIDs))])

		var question string
		var answer string

		switch game.Type {
			case model.OriginalGame:
				question = word.Original
				answer = word.Translation
			case model.TranslateGame:
				question = word.Translation
				answer = word.Original
			case model.MixedGame:
				QAs := [2]string {
					word.Original,
					word.Translation,
				}
				for i := range QAs {
					j := rand.Intn(i + 1)
					QAs[i], QAs[j] = QAs[j], QAs[i]
				}
				question = QAs[0]
				answer = QAs[1]
		}

		if err := dbConn.DB("").C(gameCollection).Update(bson.M{"userid": game.UserId}, bson.M{"$set": bson.M{"question": question,"answer": answer, "type": game.Type}}); err != nil {
			panic(fmt.Sprintf("I can't create a game for this user : %d", game.UserId))
		}

		reply.WriteString(fmt.Sprintf("Type a translation for *%s*", question))
	}

	message.Text = reply.String()
}

func GameIsStarted(id int) bool {
	return getCurrentGame(id).IsActive
}

func clearQuestionsAndAnswers(game model.Game) model.Game {
	dbConn := common.GetSession().Copy()
	defer dbConn.Close()

	if err := dbConn.DB("").C(gameCollection).Update(bson.M{"userid": game.UserId}, bson.M{"$set": bson.M{"question": "","answer": ""}}); err != nil {
		panic(fmt.Sprintf("I can't update a game for this user : %d", game.UserId))
	}
	return getCurrentGame(game.UserId)
}

func getCurrentGame(id int) model.Game {
	var currentGame model.Game
	dbConn := common.GetSession().Copy()
	defer dbConn.Close()

	dbConn.DB("").C(gameCollection).Find(bson.M{"userid": id, "isactive": true}).One(&currentGame)

	return currentGame
}

