package telegrambot

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"

	"github.com/nmaupu/nuki-logger/messaging"
	"github.com/nmaupu/nuki-logger/model"
	"github.com/rs/zerolog/log"
)

func (b *nukiBot) handlerLogs(update telego.Update, msg *telego.SendMessageParams) {
	keyboard := tu.InlineKeyboard(
		tu.InlineKeyboardRow(
			tu.InlineKeyboardButton("5").WithCallbackData(NewCallbackData("logs", "5")),
			tu.InlineKeyboardButton("10").WithCallbackData(NewCallbackData("logs", "10")),
			tu.InlineKeyboardButton("20").WithCallbackData(NewCallbackData("logs", "20")),
			tu.InlineKeyboardButton("30").WithCallbackData(NewCallbackData("logs", "30")),
			tu.InlineKeyboardButton("40").WithCallbackData(NewCallbackData("logs", "40")),
			tu.InlineKeyboardButton("50").WithCallbackData(NewCallbackData("logs", "50")),
		),
	)

	msg.Text = "How many ?"
	msg.ReplyMarkup = keyboard
	msg.ProtectContent = true
}

func (b *nukiBot) callbackLogs(update telego.Update, msg *telego.SendMessageParams) {
	data := GetDataFromCallbackData(update.CallbackQuery)
	if data == "" {
		msg.Text = "Unknown data."
		return
	}

	limit, err := strconv.Atoi(data)
	if err != nil {
		msg.Text = "Unknown number of logs."
		return
	}

	lr := b.logsReader
	lr.Limit = limit
	res, err := lr.Execute()
	if err != nil {
		msg.Text = fmt.Sprintf("Unable to get logs from API, err=%v", err)
		return
	}
	slices.Reverse(res)

	var logsLines []string
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
		str, err := b.sender.FormatLogEvent(&messaging.Event{
			Log:             l,
			ReservationName: reservationName,
		})
		if err != nil {
			log.Error().Err(err).
				Str("log_id", l.ID).
				Msg("Unable to format log event")
			continue
		}
		logsLines = append(logsLines, str)
	}

	msg.Text = strings.Join(logsLines, "\n")
}
