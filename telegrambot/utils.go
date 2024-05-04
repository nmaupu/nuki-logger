package telegrambot

import (
	"strings"

	"github.com/mymmrac/telego"
)

const (
	CallbackCommandSeparator = "|"
)

func GetCommandFromCallbackData(callbackQuery *telego.CallbackQuery) string {
	if callbackQuery == nil {
		return ""
	}
	return strings.Split(callbackQuery.Data, CallbackCommandSeparator)[0]
}

func GetDataFromCallbackData(callbackQuery *telego.CallbackQuery) string {
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
