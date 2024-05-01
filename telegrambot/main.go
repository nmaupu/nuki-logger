package telegrambot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/nmaupu/nuki-logger/messaging"
	"strings"
)

const (
	CallbackCommandSeparator = "|"
)

type CommandHandler func(update tgbotapi.Update, msg *tgbotapi.MessageConfig)

type Command struct {
	Handler  CommandHandler
	Callback CommandHandler
}
type Commands map[string]Command

func GetCommandFromCallbackData(callbackQuery *tgbotapi.CallbackQuery) string {
	if callbackQuery == nil {
		return ""
	}
	return strings.Split(callbackQuery.Data, CallbackCommandSeparator)[0]
}

func GetDataFromCallbackData(callbackQuery *tgbotapi.CallbackQuery) string {
	if callbackQuery == nil {
		return ""
	}
	toks := strings.Split(callbackQuery.Data, CallbackCommandSeparator)
	if len(toks) < 2 {
		return ""
	}
	return strings.Join(toks[1:], CallbackCommandSeparator)
}

func NewCallbackData(command string, data string) string {
	return command + CallbackCommandSeparator + data
}

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
			if update.CallbackQuery == nil && (update.Message == nil || !update.Message.IsCommand()) {
				continue
			}

			var message *tgbotapi.Message
			if update.Message != nil {
				message = update.Message
			} else {
				message = update.CallbackQuery.Message
			}

			msg := tgbotapi.NewMessage(message.Chat.ID, "")
			msg.ReplyToMessageID = message.MessageID

			var handler CommandHandler
			if update.Message != nil {
				handler = c[message.Command()].Handler
			} else {
				handler = c[GetCommandFromCallbackData(update.CallbackQuery)].Callback
			}
			if handler == nil {
				msg.Text = "Unknown command."
			} else {
				handler(update, &msg)
				if update.CallbackQuery != nil {
					config := tgbotapi.CallbackConfig{CallbackQueryID: update.CallbackQuery.ID}
					bot.AnswerCallbackQuery(config)
				}
			}
			bot.Send(msg)
		}
	}()
	return nil
}
