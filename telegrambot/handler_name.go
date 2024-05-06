package telegrambot

import (
	"context"
	"fmt"

	"github.com/looplab/fsm"
	"github.com/mymmrac/telego"
	"github.com/rs/zerolog/log"
)

func (bot nukiBot) fsmNameCommand() *fsm.FSM {
	return fsm.NewFSM(
		"idle",
		fsm.Events{
			{Name: FSMEventDefault, Src: []string{"idle"}, Dst: "waiting_for_user_input"},
			{Name: "user_input_receive", Src: []string{"waiting_for_user_input"}, Dst: "finished"},
			{Name: "reset", Src: []string{"idle", "waiting_for_user_input", "finished"}, Dst: "idle"},
		},
		fsm.Callbacks{
			FSMEventDefault: func(ctx context.Context, e *fsm.Event) {
				log.Trace().Str("callback", "run").Msg("Callback called")
				msg := &telego.SendMessageParams{}
				e.FSM.SetMetadata(FSMMetadataMessage, msg)

				msg.ParseMode = telego.ModeMarkdown
				msg.Text = "What's your *name* ?"

				waitForUserInput(e.FSM, "user_input_receive")
			},
			"user_input_receive": func(ctx context.Context, e *fsm.Event) {
				log.Trace().Str("callback", "user_input_receive").Msg("Callback called")
				userInputReceived(e.FSM)

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

				msg.ParseMode = telego.ModeMarkdown
				msg.Text = fmt.Sprintf("Hello *%s*", data)
			},
			"finished": finishedFunc,
		},
	)
}
