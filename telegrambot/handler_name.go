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
			{Name: "run", Src: []string{"idle"}, Dst: "waiting_for_user_input"},
			{Name: "user_input_receive", Src: []string{"waiting_for_user_input"}, Dst: "finished"},
			{Name: "reset", Src: []string{"finished"}, Dst: "idle"},
		},
		fsm.Callbacks{
			"run": func(ctx context.Context, e *fsm.Event) {
				log.Debug().Str("callback", "run").Msg("Callback called")
				msg := &telego.SendMessageParams{}
				e.FSM.SetMetadata(FSMMetadataMessage, msg)

				msg.ParseMode = telego.ModeMarkdown
				msg.Text = "What's your *name* ?"

				cmdI, ok := e.FSM.Metadata(FSMMetadataCommand)
				if !ok {
					log.Error().Msg("Unable to get command from metadata")
					return
				}
				cmd := cmdI.(*Command)
				cmd.NextFSMEvent = "user_input_receive"
				log.Debug().Str("func", "run").Msgf("cmd = %p", cmd)
			},
			"user_input_receive": func(ctx context.Context, e *fsm.Event) {
				log.Debug().Str("callback", "user_input_receive").Msg("Callback called")
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

				cmdI, ok := e.FSM.Metadata(FSMMetadataCommand)
				if !ok {
					log.Error().Msg("Unable to get command from metadata")
					return
				}
				cmd := cmdI.(*Command)
				cmd.NextFSMEvent = ""
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
