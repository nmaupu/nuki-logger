package telegrambot

import (
	"context"
	"errors"
	"fmt"

	"github.com/looplab/fsm"
	"github.com/mymmrac/telego"
	"github.com/rs/zerolog/log"
)

const (
	FSMEventDefault            = "run"
	FSMEventReset              = "reset"
	FSMMetadataNextEvent       = "next_event"
	FSMMetadataMessage         = "msg"
	FSMMetadataErrRecoverEvent = "err_recover_event"
)

var (
	fsmEventFinished = func(ctx context.Context, e *fsm.Event) {
		log.Debug().Str("callback", "finished").Msg("Callback called")
		if err := e.FSM.Event(ctx, "reset"); err != nil {
			log.Error().Err(err).Msg("Cannot reset")
		}
	}
)

func metadataNotFoundErr(key string) error {
	return fmt.Errorf("unable to find metadata %s", key)
}

func getMetadataReservationPendingModification(key string, fsm *fsm.FSM) (*ReservationPendingModification, error) {
	res, ok := fsm.Metadata(key)
	if !ok {
		return nil, metadataNotFoundErr(key)
	}
	m, ok := res.(*ReservationPendingModification)
	if !ok {
		return nil, metadataNotFoundErr(key)
	}
	return m, nil
}

func getMetadataString(key string, fsm *fsm.FSM) (string, error) {
	res, ok := fsm.Metadata(key)
	if !ok {
		return "", metadataNotFoundErr(key)
	}
	str, ok := res.(string)
	if !ok {
		return "", metadataNotFoundErr(key)
	}
	return str, nil
}

func getMetadataSendMessageParams(key string, fsm *fsm.FSM) (*telego.SendMessageParams, error) {
	res, ok := fsm.Metadata(key)
	if !ok {
		return nil, metadataNotFoundErr(key)
	}
	msg, ok := res.(*telego.SendMessageParams)
	if !ok {
		return nil, metadataNotFoundErr(key)
	}
	return msg, nil
}

func reinitMetadataMessage(fsm *fsm.FSM) *telego.SendMessageParams {
	msg := &telego.SendMessageParams{}
	fsm.SetMetadata(FSMMetadataMessage, msg)
	return msg
}

func checkFSMArg(e *fsm.Event) (string, error) {
	if len(e.Args) != 1 {
		return "", errors.New("invalid data")
	}

	argData := e.Args[0]
	data, ok := argData.(string)
	if !ok {
		return "", errors.New("data is not a string")
	}

	if data == "" {
		return "", errors.New("unknown data")
	}

	return data, nil
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

func fsmRuntimeErr(e *fsm.Event, err error, recoverEvent string) {
	e.Err = err
	e.FSM.SetMetadata(FSMMetadataErrRecoverEvent, recoverEvent)
}
