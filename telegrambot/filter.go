package telegrambot

import (
	"github.com/mymmrac/telego"
)

type FilterFunc func(update telego.Update) bool
