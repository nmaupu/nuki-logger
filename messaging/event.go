package messaging

import (
	"fmt"
	"nuki-logger/model"
	"strings"
	"time"
)

type Sender interface {
	Send(e *Event) error
}

type Event struct {
	Prefix string
	Log    model.NukiSmartlockLogResponse
}

func (e Event) GetValues(includeDate bool) map[string]string {
	values := map[string]string{
		"action":  e.Log.Action.String(),
		"trigger": e.Log.Trigger.String(),
		"state":   e.Log.State.String(),
		"source":  e.Log.Source.String(),
	}

	if e.Log.Source == model.NukiSourceKeypadCode {
		values["name"] = e.Log.Name
	}

	if includeDate {
		values["date"] = e.Log.Date.Format(time.RFC3339)
	}

	return values
}

func (e Event) String(includeDate bool) string {
	values := e.GetValues(includeDate)
	var valuesStr []string
	for k, v := range values {
		valuesStr = append(valuesStr, fmt.Sprintf("%s=%s", k, v))
	}
	return fmt.Sprintf("%s%s", e.Prefix, strings.Join(valuesStr, ", "))
}
