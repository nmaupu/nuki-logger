package telegrambot

import (
	"context"
	"fmt"

	"github.com/looplab/fsm"
	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
	"github.com/rs/zerolog/log"
)

func (bot nukiBot) fsmCodeConversation() *fsm.FSM {
	return fsm.NewFSM(
		"idle",
		fsm.Events{
			{Name: "run", Src: []string{"idle"}, Dst: "waiting_for_resa"},
			{Name: "resa_received", Src: []string{"idle", "waiting_for_resa"}, Dst: "finished"},
			{Name: "reset", Src: []string{"finished"}, Dst: "idle"},
		},
		fsm.Callbacks{
			"run": func(ctx context.Context, e *fsm.Event) {
				log.Debug().Str("callback", "ask_resa").Msg("Callback called")
				msgI, ok := e.FSM.Metadata("msg")
				if !ok {
					log.Error().Msg("Unable to get msg from metadata")
					return
				}
				msg := msgI.(*telego.SendMessageParams)

				res, err := bot.ReservationsReader.Execute()
				if err != nil {
					msg.Text = fmt.Sprintf("Unable to get reservations from API, err=%v", err)
					return
				}

				var keyboardButtons []telego.InlineKeyboardButton
				for _, resa := range res {
					keyboardButtons = append(keyboardButtons,
						tu.InlineKeyboardButton(fmt.Sprintf("%s (%s)", resa.Name, resa.Reference)).
							WithCallbackData(NewCallbackData("resa_received", resa.Reference)))
				}

				msg.ReplyMarkup = tu.InlineKeyboard(keyboardButtons)
				msg.Text = "Select a reservation"
				msg.ProtectContent = true
			},
			"before_resa_received": func(ctx context.Context, e *fsm.Event) {
				log.Debug().Str("callback", "resa_received").Msg("Callback called")
				msgI, ok := e.FSM.Metadata("msg")
				if !ok {
					log.Error().Msg("Unable to get msg from metadata")
					return
				}
				msg := msgI.(*telego.SendMessageParams)

				if len(e.Args) != 1 {
					msg.Text = "Invalid data."
					return
				}
				data := e.Args[0]
				if data == "" {
					msg.Text = "Unknown data."
					return
				}

				res, err := bot.SmartlockAuthReader.Execute()
				if err != nil {
					msg.Text = fmt.Sprintf("Unable to get smartlock auth from API, err=%v", err)
					return
				}
				for _, v := range res {
					if v.Name == data {
						msg.ParseMode = telego.ModeMarkdown
						msg.Text = fmt.Sprintf("Code for *%s*: %d", v.Name, v.Code)
						return
					}
				}
				msg.ParseMode = telego.ModeMarkdown
				msg.Text = fmt.Sprintf("Unable to find any code for *%s*", data)
			},
			"finished": func(ctx context.Context, e *fsm.Event) {
				log.Debug().Str("callback", "finished").Msg("Callback called")
				if err := e.FSM.Event(ctx, "reset"); err != nil {
					log.Error().Err(err).Msg("Cannot reset")
				}
			},
		},
	)
}
