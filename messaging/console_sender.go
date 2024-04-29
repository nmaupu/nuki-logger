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
		var bytes []byte
		var err error
		if e.IsLogEvent() {
			bytes, err = json.Marshal(e.Log)
		} else if e.IsSmartlockEvent() {
			bytes, err = json.Marshal(e.Smartlock)
		} else {
			err = fmt.Errorf("unable to determine the event type to send")
		}
		if err != nil {
			return err
		}
		fmt.Println(string(bytes))
		return nil
	}

	// Regular output
	if e.IsLogEvent() {
		values := e.GetValues(c.IncludeDate, false, c.Timezone)
		logger := log.With().Logger()
		for k, v := range values {
			logger = logger.With().Str(k, v).Logger()
		}
		logger.Info().Send()
	} else if e.IsSmartlockEvent() {
		log.Info().
			Str("name", e.Smartlock.Name).
			Bool("battery_critical", e.Smartlock.State.BatteryCritical).
			Bool("keypad_battery_critical", e.Smartlock.State.KeypadBatteryCritical).
			Bool("doorsensor_battery_critical", e.Smartlock.State.DoorsensorBatteryCritical).
			Int32("battery_percent", e.Smartlock.State.BatteryCharge).
			Send()
	} else {
		return fmt.Errorf("unable to determine the event type to send")
	}
	return nil
}
