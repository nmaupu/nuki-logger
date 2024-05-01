package telegrambot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/nmaupu/nuki-logger/messaging"
)

type CommandHandler func(msg *tgbotapi.MessageConfig)

type Commands map[string]CommandHandler

func (c Commands) Start(sender *messaging.TelegramSender) error {
	bot, err := tgbotapi.NewBotAPI(sender.Token)
	if err != nil {
		return err
	}
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		return err
	}
	go func() {
		for update := range updates {
			if update.Message == nil {
				continue
			}
			if !update.Message.IsCommand() {
				continue
			}

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
			msg.ReplyToMessageID = update.Message.MessageID

			commandFunc, ok := c[update.Message.Command()]
			if !ok {
				msg.Text = "Unknown command."
			} else {
				commandFunc(&msg)
			}
			bot.Send(msg)
		}
	}()
	return nil
}
