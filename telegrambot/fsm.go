package telegrambot

import (
	"context"
	"fmt"

	"github.com/looplab/fsm"
	"github.com/mymmrac/telego"
	"github.com/rs/zerolog/log"
)

const (
	FSMEventDefault      = "run"
	FSMEventReset        = "reset"
	FSMMetadataNextEvent = "next_event"
	FSMMetadataMessage   = "msg"
)

var (
	finishedFunc = func(ctx context.Context, e *fsm.Event) {
		log.Debug().Str("callback", "finished").Msg("Callback called")
		if err := e.FSM.Event(ctx, "reset"); err != nil {
			log.Error().Err(err).Msg("Cannot reset")
		}
	}
)

type FSM struct {
	*fsm.FSM
}

func metadataNotFoundErr(key string) error {
	return fmt.Errorf("unable to find metadata %s", key)
}

func (f FSM) getMetadataString(key string) (string, error) {
	res, ok := f.FSM.Metadata(key)
	if !ok {
		return "", metadataNotFoundErr(key)
	}
	str, ok := res.(string)
	if !ok {
		return "", metadataNotFoundErr(key)
	}
	return str, nil
}

func (f FSM) getMetadataSendMessageParams(key string) (*telego.SendMessageParams, error) {
	res, ok := f.FSM.Metadata(key)
	if !ok {
		return nil, metadataNotFoundErr(key)
	}
	msg, ok := res.(*telego.SendMessageParams)
	if !ok {
		return nil, metadataNotFoundErr(key)
	}
	return msg, nil
}

func waitForUserInput(f *fsm.FSM, nextEvent string) {
	f.SetMetadata(FSMMetadataNextEvent, nextEvent)
}

func userInputReceived(f *fsm.FSM) {
	userInputReset(f)
}

func userInputReset(f *fsm.FSM) {
	f.SetMetadata(FSMMetadataNextEvent, "")
}
