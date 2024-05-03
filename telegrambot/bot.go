package telegrambot

import (
	"fmt"
	"slices"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/nmaupu/nuki-logger/messaging"
	"github.com/nmaupu/nuki-logger/model"
	"github.com/nmaupu/nuki-logger/nukiapi"
	"github.com/rs/zerolog/log"
)

type NukiBot interface {
	Start() error
}

type nukiBot struct {
	sender              *messaging.TelegramSender
	logsReader          nukiapi.LogsReader
	smartlockReader     nukiapi.SmartlockReader
	reservationsReader  nukiapi.ReservationsReader
	smartlockAuthReader nukiapi.SmartlockAuthReader
}

func NewNukiBot(sender *messaging.TelegramSender,
	logsReader nukiapi.LogsReader,
	smartlockReader nukiapi.SmartlockReader,
	reservationsReader nukiapi.ReservationsReader,
	smartlockAuthReader nukiapi.SmartlockAuthReader) NukiBot {

	return &nukiBot{
		sender:              sender,
		logsReader:          logsReader,
		smartlockReader:     smartlockReader,
		reservationsReader:  reservationsReader,
		smartlockAuthReader: smartlockAuthReader,
	}

}

func (b *nukiBot) Start() error {
	commandNames := []string{
		"/help",
		"/battery",
		"/code",
		"/logs",
		"/resa",
	}
	commands := Commands{}
	commands["help"] = Command{
		Handler: func(update tgbotapi.Update, msg *tgbotapi.MessageConfig) {
			msg.Text = fmt.Sprintf("The following commands are available: %s", strings.Join(commandNames, ", "))
		},
	}

	fBattery := func(update tgbotapi.Update, msg *tgbotapi.MessageConfig) {
		res, err := b.smartlockReader.Execute()
		if err != nil {
			msg.Text = fmt.Sprintf("Unable to read smartlock status from API, err=%v", err)
		} else {
			msg.Text = res.PrettyFormat()
		}
	}
	commands["battery"] = Command{Handler: fBattery}
	commands["bat"] = Command{Handler: fBattery}
	commands["resa"] = Command{
		Handler: func(update tgbotapi.Update, msg *tgbotapi.MessageConfig) {
			res, err := b.reservationsReader.Execute()
			if err != nil {
				msg.Text = fmt.Sprintf("Unable to get reservations from API, err=%v", err)
				return
			}

			now := time.Now()
			var lines []string
			for _, r := range res {
				isBold := now.After(r.StartDate) && now.Before(r.EndDate)
				loc, err := time.LoadLocation(b.sender.Timezone)
				if err != nil {
					loc = time.UTC
				}
				startDate := r.StartDate.In(loc).Format("02/01 15:04")
				endDate := r.EndDate.In(loc).Format("02/01 15:04")
				line := fmt.Sprintf("%s (%s) - %s -> %s", r.Name, r.Reference, startDate, endDate)
				if isBold {
					line = "*" + line + "*"
				}
				lines = append(lines, line)
			}
			msg.ParseMode = tgbotapi.ModeMarkdown
			msg.Text = fmt.Sprintf("Reservations:\n%s", strings.Join(lines, "\n"))
		},
	}
	commands["logs"] = Command{
		Handler: func(update tgbotapi.Update, msg *tgbotapi.MessageConfig) {
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
		},
	}
	commands["code"] = Command{
		Handler: func(update tgbotapi.Update, msg *tgbotapi.MessageConfig) {
			res, err := b.reservationsReader.Execute()
			if err != nil {
				msg.Text = fmt.Sprintf("Unable to get reservations from API, err=%v", err)
				return
			}

			var keyboardButtons []tgbotapi.InlineKeyboardButton
			for _, r := range res {
				keyboardButtons = append(keyboardButtons,
					tgbotapi.NewInlineKeyboardButtonData(
						fmt.Sprintf("%s (%s)", r.Name, r.Reference),
						NewCallbackData("code", r.Reference)))
			}

			msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(keyboardButtons)
			msg.Text = "Select a reservation"
		},
		Callback: func(update tgbotapi.Update, msg *tgbotapi.MessageConfig) {
			data := GetDataFromCallbackData(update.CallbackQuery)
			if data == "" {
				msg.Text = "Unknown data"
				return
			}
			res, err := b.smartlockAuthReader.Execute()
			if err != nil {
				msg.Text = fmt.Sprintf("Unable to get smartlock auth from API, err=%v", err)
			}
			for _, v := range res {
				if v.Name == data {
					msg.Text = fmt.Sprintf("code: %d", v.Code)
					return
				}
			}
			msg.Text = fmt.Sprintf("Unable to find code for %s", data)
		},
	}

	return commands.start(b)
}
