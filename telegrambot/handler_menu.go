package telegrambot

import (
	"fmt"

	"github.com/enescakir/emoji"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

var (
	menuBattery = fmt.Sprintf("%s %s", emoji.Battery.String(), "battery")
	menuCode    = fmt.Sprintf("%s %s", emoji.InputNumbers.String(), "code")
	menuHelp    = fmt.Sprintf("%s %s", emoji.QuestionMark.String(), "help")
	menuLogs    = fmt.Sprintf("%s %s", emoji.FileFolder.String(), "logs")
	menuResa    = fmt.Sprintf("%s %s", emoji.OpenBook.String(), "resa")
)

func (b *nukiBot) handlerMenu(update tgbotapi.Update, msg *tgbotapi.MessageConfig) {
	numericKeyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(menuBattery),
			tgbotapi.NewKeyboardButton(menuLogs),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(menuCode),
			tgbotapi.NewKeyboardButton(menuResa),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(menuHelp),
		),
	)
	msg.Text = "Menu"
	msg.ReplyMarkup = numericKeyboard
}
