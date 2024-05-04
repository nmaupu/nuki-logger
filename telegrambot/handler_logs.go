package telegrambot

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/nmaupu/nuki-logger/messaging"
	"github.com/nmaupu/nuki-logger/model"
	"github.com/rs/zerolog/log"
)

func (b *nukiBot) handlerLogs(update tgbotapi.Update, msg *tgbotapi.MessageConfig) {
	keyboardButtons := []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("5", NewCallbackData("logs", "5")),
		tgbotapi.NewInlineKeyboardButtonData("10", NewCallbackData("logs", "10")),
		tgbotapi.NewInlineKeyboardButtonData("20", NewCallbackData("logs", "20")),
		tgbotapi.NewInlineKeyboardButtonData("30", NewCallbackData("logs", "30")),
		tgbotapi.NewInlineKeyboardButtonData("40", NewCallbackData("logs", "40")),
		tgbotapi.NewInlineKeyboardButtonData("50", NewCallbackData("logs", "50")),
	}

	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(keyboardButtons)
	msg.Text = "How many ?"
}

func (b *nukiBot) callbackLogs(update tgbotapi.Update, msg *tgbotapi.MessageConfig) {
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
