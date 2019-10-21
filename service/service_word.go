package service

import (
	"bytes"
	"fmt"
	"github.com/anyrkov/translation-memorizer/common"
	"github.com/anyrkov/translation-memorizer/model"
	"gopkg.in/mgo.v2/bson"
	"gopkg.in/telegram-bot-api.v4"
	"strings"
)

const wordsCollection = "words"

func IsExisted(update tgbotapi.Update, existedWord *model.Word) bool {
	wordIDs := GetAllUserWords(update.Message.From.ID)
	var original string

	if strings.Contains(update.Message.Text, "-") {
		original = strings.TrimSpace(strings.Split(update.Message.Text, "-")[0])
	} else {
		original = update.Message.Text
	}

	dbConn := common.GetSession().Copy()
	defer dbConn.Close()

	for _, wordId := range wordIDs {
		err := dbConn.DB("").C(wordsCollection).Find(bson.M{"id": wordId, "original": original}).One(&existedWord)
		if err == nil {
			return true
		}
	}
	return false
}

func CreateWord(update tgbotapi.Update, word *model.Word, message *tgbotapi.MessageConfig) {
	word.ID = update.Message.MessageID

	if strings.Contains(update.Message.Text, "-") {
		text := strings.Split(update.Message.Text, "-")
		word.Original = strings.TrimSpace(text[0])
		word.Translation = strings.TrimSpace(text[1])
		word.IsTranslated = true
		message.Text = fmt.Sprintf("Done! original: *%s* - translation: *%s*!", word.Original, word.Translation)
		message.ParseMode = "markdown"
	} else {
		word.Original = update.Message.Text
		word.IsTranslated = false
		word.Translation = ""
		message.Text = fmt.Sprintf("Ok and now give me a translation for *%s*!", word.Original)
	}

	dbConn := common.GetSession().Copy()
	defer dbConn.Close()

	if err := dbConn.DB("").C(wordsCollection).Insert(word); err != nil {
		panic(fmt.Sprintf("I can't insert this word : %s", word.Original))
	}
}

func AddTranslation(word model.Word, update tgbotapi.Update) {
	dbConn := common.GetSession().Copy()
	defer dbConn.Close()

	if err := dbConn.DB("").C(wordsCollection).Update(bson.M{"id": word.ID}, bson.M{"$set": bson.M{"translation": update.Message.Text,"istranslated": true}}); err != nil {
		panic(fmt.Sprintf("I can't insert this translation : %s", word.Translation))
	}

}

func ShowAllWords(update tgbotapi.Update, message *tgbotapi.MessageConfig) {
	wordIDs := GetAllUserWords(update.Message.From.ID)

	if len(wordIDs) == 0 {
		message.Text = "You don't have a words"
		return
	}

	dbConn := common.GetSession().Copy()
	defer dbConn.Close()

	var word model.Word
	var words []model.Word
	for _, id := range wordIDs {
		if err := dbConn.DB("").C(wordsCollection).Find(bson.M{"id": id}).One(&word); err == nil {
			words = append(words, word)
		}
	}

	var text bytes.Buffer

	for _, word := range words {
		text.WriteString(fmt.Sprintf(
			"%s - %s \n",
			word.Original,
			word.Translation))
	}

	message.Text = text.String()
}

func HasWordWithoutTranslation(userId int, word *model.Word) bool {
	wordIDs := GetAllUserWords(userId)
	dbConn := common.GetSession().Copy()
	defer dbConn.Close()

	for _, wordId := range wordIDs {
		err := dbConn.DB("").C(wordsCollection).Find(bson.M{"id": wordId, "istranslated": false}).One(&word)
		if err == nil {
			return true
		}
	}
	return false
}

func ShowAllWordsForDelete(update tgbotapi.Update, message *tgbotapi.MessageConfig) {
	keyboard := tgbotapi.InlineKeyboardMarkup{}

	wordIDs := GetAllUserWords(update.Message.From.ID)

	dbConn := common.GetSession().Copy()
	defer dbConn.Close()

	if len(wordIDs) == 0 {
		message.Text = "You don't have a words"
		return
	}

	var word model.Word
	var wordNames []string
	for _, id := range wordIDs {
		if err := dbConn.DB("").C(wordsCollection).Find(bson.M{"id": id}).One(&word); err == nil {
			wordNames = append(wordNames, word.Original)
		}
	}

	for _, wordName := range wordNames {
		var row []tgbotapi.InlineKeyboardButton
		btn := tgbotapi.NewInlineKeyboardButtonData(wordName, wordName)
		row = append(row, btn)
		keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, row)
	}

	message.ReplyMarkup = keyboard
	message.Text = "Select the word you want to delete"
}

func DeleteWord(userId int, word model.Word) {
	deleteWord(word)
	deleteWordFromUser(userId, word.ID)
}

func deleteWord(word model.Word) {
	dbConn := common.GetSession().Copy()
	defer dbConn.Close()

	if err := dbConn.DB("").C(wordsCollection).Remove(word); err != nil {
		panic(fmt.Sprintf("I can't remove this word"))
	}
}

func GetWordByOriginal(original string) model.Word {
	var word model.Word
	dbConn := common.GetSession().Copy()
	defer dbConn.Close()

	if err := dbConn.DB("").C(wordsCollection).Find(bson.M{"original": original}).One(&word); err != nil {
		panic(fmt.Sprintf("The word with name - %s doesn't exist", original))
	}
	return word
}

func getWordById(id int) model.Word {
	var word model.Word
	dbConn := common.GetSession().Copy()
	defer dbConn.Close()

	if err := dbConn.DB("").C(wordsCollection).Find(bson.M{"id": id}).One(&word); err != nil {
		panic(fmt.Sprintf("The word with id - %d doesn't exist", id))
	}
	return word
}

