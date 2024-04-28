package messaging

import (
	"github.com/rs/zerolog/log"
)

var (
	_ Sender = (*ConsoleSender)(nil)
)

type ConsoleSender struct {
	sender
}

func (c *ConsoleSender) Send(e *Event) error {
	values := e.GetValues(c.IncludeDate)
	logger := log.With().Logger()
	for k, v := range values {
		logger = logger.With().Str(k, v).Logger()
	}
	logger.Info().Send()
	return nil
}
