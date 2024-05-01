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
	var err error
	var msg string

	botAPI, err := tgbotapi.NewBotAPI(t.Token)
	if err != nil {
		return err
	}

	if e.IsLogEvent() {
		msg, err = t.processLogEvent(e)
		if err != nil {
			return err
		}
	} else if e.IsSmartlockEvent() {
		msg, err = t.processSmartlockEvent(e)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("unable to determine the type of event to send")
	}

	_, err = botAPI.Send(tgbotapi.NewMessage(t.ChatID, msg))
	return err
}

func (t *TelegramSender) processLogEvent(e *Event) (string, error) {
	if e.Json {
		bytes, err := json.Marshal(e.Log)
		if err != nil {
			return "", err
		}

		return string(bytes), nil
	}

	var date string
	if t.IncludeDate {
		loc, err := time.LoadLocation(t.Timezone)
		if err != nil {
			loc = time.UTC
		}
		date = e.Log.Date.In(loc).Format(time.DateTime) + " - "
	}

	switch {
	case e.Log.Trigger == model.NukiTriggerButton:
		// Lock / unlock with button
		return fmt.Sprintf("%s%s %s %s",
			date, e.Log.Trigger.GetEmoji(), e.Log.Action.String(), e.Log.State.GetEmoji()), nil
	case e.Log.Trigger == model.NukiTriggerKeypad && e.Log.Source == model.NukiSourceKeypadCode:
		// Someone enters keypad code
		name := e.ReservationName
		if name == "" {
			name = e.Log.Name
		}
		return fmt.Sprintf("%s%s %s by '%s' %s",
			date, e.Log.Trigger.GetEmoji(), e.Log.Action.String(), e.ReservationName, e.Log.State.GetEmoji()), nil
	case e.Log.Trigger == model.NukiTriggerKeypad && e.Log.Source == model.NukiSourceDefault:
		// < keypad button is pressed
		return fmt.Sprintf("%s%s %s %s",
			date, e.Log.Trigger.GetEmoji(), e.Log.Action.String(), e.Log.State.GetEmoji()), nil
	case e.Log.Trigger == model.NukiTriggerSystem && (e.Log.Action == model.NukiActionDoorOpened || e.Log.Action == model.NukiActionDoorClosed):
		return fmt.Sprintf("%s%s %s %s",
			date, emoji.Door.String(), e.Log.Action.String(), e.Log.State.GetEmoji()), nil
	case e.Log.Trigger == model.NukiTriggerManual:
		return fmt.Sprintf("%s%s %s %s",
			date, e.Log.Trigger.GetEmoji(), e.Log.Action.String(), e.Log.State.GetEmoji()), nil
	default:
		return e.String(t.IncludeDate, true, t.Timezone), nil
	}
}

func (t *TelegramSender) processSmartlockEvent(e *Event) (string, error) {
	if e.Json {
		bytes, err := json.Marshal(e.Smartlock.ToSmartlockState())
		if err != nil {
			return "", err
		}
		return string(bytes), nil
	}

	return e.Smartlock.PrettyFormat(), nil
}
