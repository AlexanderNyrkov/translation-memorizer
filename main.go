package main

import (
	"fmt"
	"github.com/anyrkov/translation-memorizer/common"
	"github.com/anyrkov/translation-memorizer/model"
	"github.com/anyrkov/translation-memorizer/service"
	"gopkg.in/telegram-bot-api.v4"
	"log"
	"os"
	"path/filepath"
)

func main() {
	const configDir = "config"
	baseDir, _ := os.Getwd()
	tokenPath := filepath.Join(baseDir, configDir)
	token, _ := common.ReadToken(filepath.Join(tokenPath, "token.json"))

	bot, err := tgbotapi.NewBotAPI(token)

	if err != nil {
		log.Panic(err)
	}

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	var msg tgbotapi.MessageConfig

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {

		if update.Message == nil {
			if update.CallbackQuery != nil {
				service.DeleteWord(update.CallbackQuery.From.ID, service.GetWordByOriginal(update.CallbackQuery.Data))
				msg.Text = fmt.Sprintf("*%s* was deleted", update.CallbackQuery.Data)
				msg.ReplyMarkup = tgbotapi.ReplyKeyboardRemove{RemoveKeyboard: true}
				bot.Send(msg)
			}
			continue
		}

		var user model.User

		msg.ReplyMarkup = tgbotapi.ReplyKeyboardRemove{RemoveKeyboard: true}
		msg.ChatID = update.Message.Chat.ID
		service.CreateUser(update.Message.From.ID, &user)

		command := update.Message.Command()

		if command != "" {
			switch command {
			case common.ShowAllWords.String():
				service.ShowAllWords(update, &msg)
			case common.StartGame.String():
				service.StartGame(update, &msg)
			case common.StopGame.String():
				service.StopGame(update, &msg)
			case common.DeleteWord.String():
				service.ShowAllWordsForDelete(update, &msg)
			case common.Help.String():
				msg.ParseMode = "markdown"
				msg.Text = `
						*Words*

*/show_me_all_words* - shows all the words with translations that you added
*/delete_word*       - shows all words and you can choose which one to delete
						
if you want to add a word then just write it and then write a translation for it
or you can use a hyphen:
_Original - Translation_

*Games*

*/stop_game*         - stops a game
*/start_game*        - starts a game:
_original game_      - you should write a translation of original word
_translation game_   - you should write an original of translation word
_mixed game_         - mixed original and translation game
`
			}
			bot.Send(msg)
			continue
		}

		msg.ParseMode = "markdown"

		if service.GameIsStarted(update.Message.From.ID) {
			service.PlayGame(update, &msg)
			bot.Send(msg)
			continue
		}

		var word model.Word

		if service.HasWordWithoutTranslation(user.ID, &word) {
			service.AddTranslation(word, update)
			msg.Text = fmt.Sprintf("Added a translation for *%s* !", word.Original)
			bot.Send(msg)
			continue
		}

		if service.IsExisted(update, &word) {
			msg.Text = fmt.Sprintf("This word already exist! : *%s* - *%s*", word.Original, word.Translation)
			bot.Send(msg)
			continue
		}
		service.CreateWord(update, &word, &msg)
		service.AddWordToUser(user.ID, word.ID)

		bot.Send(msg)
	}
}