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

			var destinationChatID int64
			if update.Message != nil && !isPrivateMessage(update) {
				// Command are only executed through private messages, deleting message.
				log.Debug().Msg("Ignoring commands sent to group")
				_, err := bot.DeleteMessage(tgbotapi.NewDeleteMessage(update.Message.Chat.ID, update.Message.MessageID))
				if err != nil {
					log.Error().Err(err).
						Int64("chat_id", update.Message.Chat.ID).
						Int("message_id", update.Message.MessageID).
						Str("message", update.Message.Text).
						Msg("Unable to delete unwanted message")
				}
				// if it's a command: answer response to the member
				destinationChatID = int64(update.Message.From.ID)
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
					command = "start"
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

			if destinationChatID == 0 {
				destinationChatID = message.Chat.ID
			}
			msgToSend := tgbotapi.NewMessage(destinationChatID, "")
			msgToSend.ReplyToMessageID = 0
			msgToSend.ParseMode = tgbotapi.ModeMarkdown

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
