package telegrambot

import (
	"fmt"

	"github.com/mymmrac/telego"
)

func (b *nukiBot) handlerSavePendingReservationsToCache(update telego.Update, msg *telego.SendMessageParams) {
	if err := b.reservationPendingModificationRoutine.SaveToCache(); err != nil {
		msg.Text = fmt.Sprintf("Unable to save to cache, err=%v", err)
	} else {
		msg.Text = "Done."
	}
}
