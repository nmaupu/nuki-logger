package telegrambot

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/looplab/fsm"
	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"

	"github.com/nmaupu/nuki-logger/messaging"
	"github.com/nmaupu/nuki-logger/model"
	"github.com/rs/zerolog/log"
)

func (bot nukiBot) fsmLogsCommand() *fsm.FSM {
	return fsm.NewFSM(
		"idle",
		fsm.Events{
			{Name: FSMEventDefault, Src: []string{"idle"}, Dst: "wait_for_number"},
			{Name: "number_received", Src: []string{"idle", "wait_for_number"}, Dst: "finished"},
			{Name: "reset", Src: []string{"idle", "wait_for_number", "finished"}, Dst: "idle"},
		},
		fsm.Callbacks{
			FSMEventDefault: func(ctx context.Context, e *fsm.Event) {
				log.Trace().Str("callback", "run").Msg("Callback called")
				msg := &telego.SendMessageParams{}
				e.FSM.SetMetadata(FSMMetadataMessage, msg)

				keyboard := tu.InlineKeyboard(
					tu.InlineKeyboardRow(
						tu.InlineKeyboardButton("5").WithCallbackData(NewCallbackData("number_received", "5")),
						tu.InlineKeyboardButton("10").WithCallbackData(NewCallbackData("number_received", "10")),
						tu.InlineKeyboardButton("20").WithCallbackData(NewCallbackData("number_received", "20")),
						tu.InlineKeyboardButton("30").WithCallbackData(NewCallbackData("number_received", "30")),
						tu.InlineKeyboardButton("40").WithCallbackData(NewCallbackData("number_received", "40")),
						tu.InlineKeyboardButton("50").WithCallbackData(NewCallbackData("number_received", "50")),
					),
				)

				msg.Text = "How many ?"
				msg.ReplyMarkup = keyboard
				msg.ProtectContent = true
			},
			"before_number_received": func(ctx context.Context, e *fsm.Event) {
				log.Trace().Str("callback", "number_received").Msg("Callback called")

				msg := &telego.SendMessageParams{}
				e.FSM.SetMetadata(FSMMetadataMessage, msg)

				if len(e.Args) != 1 {
					msg.Text = "Invalid data."
					return
				}
				data := e.Args[0]
				if data == "" {
					msg.Text = "Unknown data."
					return
				}

				limit, err := strconv.Atoi(data.(string))
				if err != nil {
					msg.Text = "Unknown number of logs."
					return
				}

				lr := bot.LogsReader
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
						reservationName, err = bot.ReservationsReader.GetReservationName(l.Name)
						if err != nil {
							logger.Error().
								Err(err).
								Msg("Unable to get reservation's name, keeping original ref as name")
							reservationName = l.Name
						}
					}

					str, err := bot.Sender.FormatLogEvent(&messaging.Event{
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
			},
			"finished": finishedFunc,
		},
	)
}
