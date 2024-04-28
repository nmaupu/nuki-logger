package messaging

import (
	"encoding/json"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

var (
	_ Sender = (*TelegramSender)(nil)
)

type TelegramSender struct {
	sender `mapstructure:",squash"`
	Token  string `mapstructure:"token"`
	ChatID int64  `mapstructure:"chat_id"`
}

func (t *TelegramSender) Send(e *Event) error {
	botAPI, err := tgbotapi.NewBotAPI(t.Token)
	if err != nil {
		return err
	}
	var msg string

	if e.Json {
		bytes, err := json.Marshal(e.Log)
		if err != nil {
			return err
		}
		msg = string(bytes)
	} else {
		msg = e.String(t.IncludeDate)
	}

	_, err = botAPI.Send(tgbotapi.NewMessage(t.ChatID, msg))
	return err
}
