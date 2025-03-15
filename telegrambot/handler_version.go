package telegrambot

import (
	"fmt"

	"github.com/mymmrac/telego"
	"github.com/nmaupu/nuki-logger/model"
	"github.com/rs/zerolog/log"
)

func (b *nukiBot) handlerVersion(update telego.Update, msg *telego.SendMessageParams) {
	log.Debug().Msg("handlerVersion called")

	msg.ParseMode = telego.ModeMarkdown
	msg.Text = fmt.Sprintf("%s, version %s, build %s", model.AppName, model.ApplicationVersion, model.BuildDate)
}
