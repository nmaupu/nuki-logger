package telegrambot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/rs/zerolog/log"
)

type CommandHandler func(update tgbotapi.Update, msgResponse *tgbotapi.MessageConfig)

type Command struct {
	Handler  CommandHandler
	Callback CommandHandler
}
type Commands map[string]Command

func (c Commands) start(b *nukiBot) error {
	bot, err := tgbotapi.NewBotAPI(b.sender.Token)
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
			if update.CallbackQuery == nil && update.Message == nil {
				continue
			}

			if update.Message != nil && !isPrivateMessage(update) {
				// Command are only executed through private messages, skipping.
				log.Debug().Msg("Ignoring commands sent to group")
				continue
			}

			var command string
			if update.Message != nil && !update.Message.IsCommand() {
				// Menu click
				switch update.Message.Text {
				case menuResa:
					command = "resa"
				case menuCode:
					command = "code"
				case menuLogs:
					command = "logs"
				case menuBattery:
					command = "battery"
				default:
					command = "help"
				}
			} else if update.Message != nil {
				command = update.Message.Command()
			}

			var message *tgbotapi.Message
			if update.Message != nil {
				message = update.Message
			} else {
				message = update.CallbackQuery.Message
			}

			msgToSend := tgbotapi.NewMessage(message.Chat.ID, "")
			msgToSend.ReplyToMessageID = 0

			var fn CommandHandler
			if update.Message != nil {
				fn = c[command].Handler
			} else {
				fn = c[GetCommandFromCallbackData(update.CallbackQuery)].Callback
			}
			if fn == nil {
				msgToSend.Text = "Unknown command."
			} else {
				fn(update, &msgToSend)
				if update.CallbackQuery != nil { // fn is a callback func, answering callback ok
					config := tgbotapi.CallbackConfig{CallbackQueryID: update.CallbackQuery.ID}
					_, _ = bot.AnswerCallbackQuery(config)
				}
			}
			_, _ = bot.Send(msgToSend)
		}
	}()
	return nil
}

func isPrivateMessage(update tgbotapi.Update) bool {
	return update.Message != nil && update.Message.Chat != nil && update.Message.Chat.IsPrivate()
}
