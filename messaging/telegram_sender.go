package messaging

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

var (
	_ Sender = (*TelegramSender)(nil)
)

type TelegramSender struct {
	sender
	Token  string `mapstructure:"token"`
	ChatID int64  `mapstructure:"chat_id"`
}

func (t *TelegramSender) Send(e *Event) error {
	botAPI, err := tgbotapi.NewBotAPI(t.Token)
	if err != nil {
		return err
	}
	_, err = botAPI.Send(tgbotapi.NewMessage(t.ChatID, e.String(t.IncludeDate)))
	return err
}
