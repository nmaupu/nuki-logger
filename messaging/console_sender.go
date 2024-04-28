package messaging

import (
	"github.com/rs/zerolog/log"
)

var (
	_ Sender = (*ConsoleSender)(nil)
)

type ConsoleSender struct {
	IncludeDate bool `mapstructure:"include_date"`
}

func (t *ConsoleSender) Send(e *Event) error {
	values := e.GetValues(t.IncludeDate)
	logger := log.With().Logger()
	for k, v := range values {
		logger = logger.With().Str(k, v).Logger()
	}
	logger.Info().Send()
	return nil
}
