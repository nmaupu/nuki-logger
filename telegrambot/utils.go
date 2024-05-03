package telegrambot

import (
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

const (
	CallbackCommandSeparator = "|"
)

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
