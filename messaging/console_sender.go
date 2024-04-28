package messaging

import (
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog/log"
)

var (
	_ Sender = (*ConsoleSender)(nil)
)

type ConsoleSender struct {
	sender `mapstructure:",squash"`
}

func (c *ConsoleSender) Send(e *Event) error {
	if e.Json {
		bytes, err := json.Marshal(e.Log)
		fmt.Println(string(bytes))
		return err
	}

	// Regular output
	values := e.GetValues(c.IncludeDate, false, c.Timezone)
	logger := log.With().Logger()
	for k, v := range values {
		logger = logger.With().Str(k, v).Logger()
	}
	logger.Info().Send()
	return nil
}
