package telegrambot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type CommandHandler func(update tgbotapi.Update, msg *tgbotapi.MessageConfig)

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
			var handler CommandHandler

			if update.CallbackQuery == nil && update.Message == nil {
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

			msg := tgbotapi.NewMessage(message.Chat.ID, "")
			msg.ReplyToMessageID = 0

			if update.Message != nil {
				handler = c[command].Handler
			} else {
				handler = c[GetCommandFromCallbackData(update.CallbackQuery)].Callback
			}
			if handler == nil {
				msg.Text = "Unknown command."
			} else {
				handler(update, &msg)
				if update.CallbackQuery != nil {
					config := tgbotapi.CallbackConfig{CallbackQueryID: update.CallbackQuery.ID}
					_, _ = bot.AnswerCallbackQuery(config)
				}
			}
			_, _ = bot.Send(msg)
		}
	}()
	return nil
}
