package telegrambot

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func (b *nukiBot) handlerCode(update tgbotapi.Update, msg *tgbotapi.MessageConfig) {
	res, err := b.reservationsReader.Execute()
	if err != nil {
		msg.Text = fmt.Sprintf("Unable to get reservations from API, err=%v", err)
		return
	}

	var keyboardButtons []tgbotapi.InlineKeyboardButton
	for _, r := range res {
		keyboardButtons = append(keyboardButtons,
			tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("%s (%s)", r.Name, r.Reference),
				NewCallbackData("code", r.Reference)))
	}

	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(keyboardButtons)
	msg.Text = "Select a reservation"
}

func (b *nukiBot) callbackCode(update tgbotapi.Update, msg *tgbotapi.MessageConfig) {
	data := GetDataFromCallbackData(update.CallbackQuery)
	if data == "" {
		msg.Text = "Unknown data"
		return
	}
	res, err := b.smartlockAuthReader.Execute()
	if err != nil {
		msg.Text = fmt.Sprintf("Unable to get smartlock auth from API, err=%v", err)
	}
	for _, v := range res {
		if v.Name == data {
			msg.Text = fmt.Sprintf("code for %s: %d", v.Name, v.Code)
			return
		}
	}
	msg.Text = fmt.Sprintf("Unable to find code for %s", data)
}
