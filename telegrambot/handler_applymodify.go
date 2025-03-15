package telegrambot

import (
	"github.com/rs/zerolog/log"

	"github.com/mymmrac/telego"
)

func (b *nukiBot) handlerApplyModify(update telego.Update, msg *telego.SendMessageParams) {
	log.Debug().Msg("handlerApplyModify called")
	b.reservationPendingModificationRoutine.ApplyModificationNow()
}
