package telegrambot

import (
	"context"
	"fmt"

	"github.com/looplab/fsm"
	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
	"github.com/rs/zerolog/log"
)

func (bot nukiBot) fsmCodeCommand() *fsm.FSM {
	return fsm.NewFSM(
		"idle",
		fsm.Events{
			{Name: FSMEventDefault, Src: []string{"idle"}, Dst: "waiting_for_resa"},
			{Name: "resa_received", Src: []string{"idle", "waiting_for_resa"}, Dst: "finished"},
			{Name: "reset", Src: []string{"idle", "waiting_for_resa", "finished"}, Dst: "idle"},
		},
		fsm.Callbacks{
			FSMEventDefault: func(ctx context.Context, e *fsm.Event) {
				log.Debug().Str("callback", "run").Msg("Callback called")
				msg := &telego.SendMessageParams{}
				e.FSM.SetMetadata(FSMMetadataMessage, msg)

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
				msg.ParseMode = telego.ModeMarkdown
				msg.Text = "What *reservation* do you want the code for ?"
				msg.ProtectContent = true
			},
			"before_resa_received": func(ctx context.Context, e *fsm.Event) {
				log.Debug().Str("callback", "resa_received").Msg("Callback called")
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

				res, err := bot.SmartlockAuthReader.Execute()
				if err != nil {
					msg.Text = fmt.Sprintf("Unable to get smartlock auth from API, err=%v", err)
					return
				}
				msg.ParseMode = telego.ModeMarkdown
				for _, v := range res {
					if v.Name == data {
						msg.Text = fmt.Sprintf("Code for *%s*: %d", v.Name, v.Code)
						return
					}
				}
				msg.Text = fmt.Sprintf("Unable to find any code for *%s*", data)
			},
			"finished": finishedFunc,
		},
	)
}
