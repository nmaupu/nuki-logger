package messaging

import (
	"fmt"
	"github.com/nmaupu/nuki-logger/model"
	"strings"
	"time"
)

type Sender interface {
	Send(e *Event) error
	GetName() string
}

type sender struct {
	Name        string `mapstructure:"-"`
	IncludeDate bool   `mapstructure:"include_date"`
	Timezone    string `mapstructure:"timezone"`
}

func (s *sender) GetName() string {
	return s.Name
}

type Event struct {
	Prefix string
	Log    model.NukiSmartlockLogResponse
	Json   bool
}

func (e Event) GetValues(includeDate, emoji bool, tz string) map[string]string {
	var values map[string]string
	if emoji {
		values = map[string]string{
			"action":  e.Log.Action.String(),
			"trigger": e.Log.Trigger.GetEmoji(),
			"state":   e.Log.State.GetEmoji(),
			"source":  e.Log.Source.String(),
		}
	} else {
		values = map[string]string{
			"action":  e.Log.Action.String(),
			"trigger": e.Log.Trigger.String(),
			"state":   e.Log.State.String(),
			"source":  e.Log.Source.String(),
		}
	}

	if e.Log.Source == model.NukiSourceKeypadCode {
		values["name"] = e.Log.Name
	}

	if includeDate {
		loc, err := time.LoadLocation(tz)
		if err != nil {
			loc = time.UTC
		}
		values["date"] = e.Log.Date.In(loc).Format(time.DateTime)
	}

	return values
}

func (e Event) String(includeDate, emoji bool, tz string) string {
	values := e.GetValues(includeDate, emoji, tz)
	var valuesStr []string
	for k, v := range values {
		valuesStr = append(valuesStr, fmt.Sprintf("%s=%s", k, v))
	}
	return fmt.Sprintf("%s%s", e.Prefix, strings.Join(valuesStr, ", "))
}
