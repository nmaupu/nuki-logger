package telegrambot

import (
	"fmt"
	"slices"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/nmaupu/nuki-logger/messaging"
	"github.com/nmaupu/nuki-logger/model"
	"github.com/rs/zerolog/log"
)

func (b *nukiBot) handlerLogs(update tgbotapi.Update, msg *tgbotapi.MessageConfig) {
	lr := b.logsReader
	lr.Limit = 10
	res, err := lr.Execute()
	if err != nil {
		msg.Text = fmt.Sprintf("Unable to get logs from API, err=%v", err)
		return
	}
	slices.Reverse(res)
	for _, l := range res {
		logger := log.With().
			Str("ref", l.Name).
			Str("command", "logs").
			Logger()
		reservationName := l.Name
		if l.Trigger == model.NukiTriggerKeypad && l.Source == model.NukiSourceKeypadCode && l.State != model.NukiStateWrongKeypadCode {
			reservationName, err = b.reservationsReader.GetReservationName(l.Name)
			if err != nil {
				logger.Error().
					Err(err).
					Msg("Unable to get reservation's name, keeping original ref as name")
				reservationName = l.Name
			}
		}
		err := b.sender.Send(&messaging.Event{
			Log:             l,
			ReservationName: reservationName,
		})
		if err != nil {
			logger.Error().Err(err).Msg("Unable to send message to telegram")
		}
	}
}
