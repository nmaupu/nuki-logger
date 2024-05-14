package telegrambot

import (
	"fmt"
	"strings"

	"github.com/enescakir/emoji"
	"github.com/rs/zerolog/log"

	"github.com/mymmrac/telego"
)

func (b *nukiBot) handlerListModify(update telego.Update, msg *telego.SendMessageParams) {
	log.Debug().Msg("handlerListModify called")

	modifs := b.reservationPendingModificationRoutine.GetAllPendingModifications()
	if len(modifs) == 0 {
		msg.Text = "No pending modification"
		return
	}

	strs := []string{}
	for _, v := range modifs {
		em := emoji.HourglassNotDone.String()
		if v.ModificationDone {
			em = emoji.CheckMark.String()
		}
		strs = append(strs, fmt.Sprintf("*%s*: %s - %s %s", v.ReservationRef, v.FormatCheckIn(), v.FormatCheckOut(), em))
	}

	msg.ParseMode = telego.ModeMarkdown
	msg.Text = strings.Join(strs, "\n")
}
