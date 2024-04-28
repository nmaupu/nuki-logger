package messaging

import (
	"encoding/json"
	"fmt"
	"github.com/enescakir/emoji"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/nmaupu/nuki-logger/model"
	"time"
)

var (
	_ Sender = (*TelegramSender)(nil)
)

type TelegramSender struct {
	sender `mapstructure:",squash"`
	Token  string `mapstructure:"token"`
	ChatID int64  `mapstructure:"chat_id"`
}

func (t *TelegramSender) Send(e *Event) error {
	botAPI, err := tgbotapi.NewBotAPI(t.Token)
	if err != nil {
		return err
	}
	var msg string

	if e.Json {
		bytes, err := json.Marshal(e.Log)
		if err != nil {
			return err
		}
		msg = string(bytes)
	} else {
		var date string
		if t.IncludeDate {
			loc, err := time.LoadLocation(t.Timezone)
			if err != nil {
				loc = time.UTC
			}
			date = e.Log.Date.In(loc).Format(time.DateTime) + " - "
		}
		if e.Log.Trigger == model.NukiTriggerButton {
			// Lock / unlock with button
			msg = fmt.Sprintf("%s%s %s %s", date, e.Log.Trigger.GetEmoji(), e.Log.Action.String(), e.Log.State.GetEmoji())
		} else if e.Log.Trigger == model.NukiTriggerKeypad && e.Log.Source == model.NukiSourceKeypadCode {
			// Someone enters keypad code
			msg = fmt.Sprintf("%s%s %s by '%s' %s", date, e.Log.Trigger.GetEmoji(), e.Log.Action.String(), e.Log.Name, e.Log.State.GetEmoji())
		} else if e.Log.Trigger == model.NukiTriggerKeypad && e.Log.Source == model.NukiSourceDefault {
			// < keypad button is pressed
			msg = fmt.Sprintf("%s%s %s %s", date, e.Log.Trigger.GetEmoji(), e.Log.Action.String(), e.Log.State.GetEmoji())
		} else if e.Log.Trigger == model.NukiTriggerSystem && (e.Log.Action == model.NukiActionDoorOpened || e.Log.Action == model.NukiActionDoorClosed) {
			// door opened / closed
			msg = fmt.Sprintf("%s%s %s %s", date, emoji.Door.String(), e.Log.Action.String(), e.Log.State.GetEmoji())
		} else {
			msg = e.String(t.IncludeDate, true, t.Timezone)
		}
	}

	_, err = botAPI.Send(tgbotapi.NewMessage(t.ChatID, msg))
	return err
}
