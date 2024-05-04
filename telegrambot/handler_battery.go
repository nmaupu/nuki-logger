package telegrambot

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func (b *nukiBot) handlerBattery(update tgbotapi.Update, msg *tgbotapi.MessageConfig) {
	res, err := b.smartlockReader.Execute()
	if err != nil {
		msg.Text = fmt.Sprintf("Unable to read smartlock status from API, err=%v", err)
	} else {
		msg.Text = res.PrettyFormat()
	}
}
