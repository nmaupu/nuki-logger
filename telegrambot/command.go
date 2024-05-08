package telegrambot

import (
	"strings"

	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"

	"github.com/rs/zerolog/log"
)

type CommandHandler func(update telego.Update, msgResponse *telego.SendMessageParams)

type Command struct {
	Handler  CommandHandler
	Callback CommandHandler
}
type Commands map[string]Command

func (c Commands) start(b *nukiBot) error {
	bot, err := telego.NewBot(b.sender.Token)
	if err != nil {
		return err
	}

	updates, err := bot.UpdatesViaLongPolling(nil)
	if err != nil {
		return err
	}

	go func() {
		defer bot.StopLongPolling()
	POLL:
		for update := range updates {
			// Execute all filters before proceeding
			for _, filterFunc := range b.filters {
				if !filterFunc(update) {
					logger := log.With().Logger()
					if update.Message != nil {
						logger = logger.With().
							Int64("from_id", update.Message.From.ID).
							Str("from_username", update.Message.From.Username).
							Str("from_firstname", update.Message.From.FirstName).
							Str("from_lastname", update.Message.From.LastName).
							Str("from_lang", update.Message.From.LanguageCode).
							Str("message", update.Message.Text).Logger()
					}
					logger.Warn().Msg("Message filtered.")
					continue POLL
				}
			}

			if update.CallbackQuery == nil && update.Message == nil {
				continue
			}

			var destinationChatID int64
			if update.Message != nil && !IsPrivateMessage(update) {

				// Command are only executed through private messages, deleting message.
				log.Debug().Msg("Ignoring commands sent to group")
				err := bot.DeleteMessage(tu.Delete(update.Message.Chat.ChatID(), update.Message.MessageID))
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
			if update.Message != nil && !strings.HasPrefix(update.Message.Text, "/") {
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
				command, _ = strings.CutPrefix(update.Message.Text, "/")
			}

			if destinationChatID == 0 {
				if update.Message != nil {
					destinationChatID = update.Message.Chat.ID
				} else {
					destinationChatID = update.CallbackQuery.Message.GetChat().ID
				}
			}

			msgToSend := tu.Message(tu.ID(destinationChatID), "")
			msgToSend.ParseMode = telego.ModeMarkdown

			var fn CommandHandler
			if update.Message != nil {
				fn = c[command].Handler
			} else {
				fn = c[GetCommandFromCallbackData(update.CallbackQuery)].Callback
			}
			if fn == nil {
				msgToSend.Text = "Unknown command."
			} else {
				fn(update, msgToSend)
				if update.CallbackQuery != nil { // fn is a callback func, answering callback ok
					_ = bot.AnswerCallbackQuery(tu.CallbackQuery(update.CallbackQuery.ID))
				}
			}
			_, _ = bot.SendMessage(msgToSend)
		}
	}()
	return nil
}

func IsPrivateMessage(update telego.Update) bool {
	return update.Message != nil && update.Message.Chat.Type == telego.ChatTypePrivate
}
