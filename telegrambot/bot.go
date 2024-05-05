package telegrambot

import (
	"fmt"
	"strings"

	"github.com/mymmrac/telego"
	"github.com/nmaupu/nuki-logger/messaging"
	"github.com/nmaupu/nuki-logger/nukiapi"
	sf "github.com/sa-/slicefunk"
	"golang.org/x/exp/maps"
)

type NukiBot interface {
	Start() error
}

type nukiBot struct {
	Sender              *messaging.TelegramSender
	LogsReader          nukiapi.LogsReader
	SmartlockReader     nukiapi.SmartlockReader
	ReservationsReader  nukiapi.ReservationsReader
	SmartlockAuthReader nukiapi.SmartlockAuthReader
}

func NewNukiBot(sender *messaging.TelegramSender,
	logsReader nukiapi.LogsReader,
	smartlockReader nukiapi.SmartlockReader,
	reservationsReader nukiapi.ReservationsReader,
	smartlockAuthReader nukiapi.SmartlockAuthReader) NukiBot {

	return &nukiBot{
		Sender:              sender,
		LogsReader:          logsReader,
		SmartlockReader:     smartlockReader,
		ReservationsReader:  reservationsReader,
		SmartlockAuthReader: smartlockAuthReader,
	}
}

func (b *nukiBot) Start() error {
	commands := Commands{}
	help := func(update telego.Update, msg *telego.SendMessageParams) {
		keys := sf.Map(maps.Keys(commands), func(item string) string { return "/" + item })
		msg.Text = fmt.Sprintf("The following commands are available: %s", strings.Join(keys, ", "))
	}
	commands["start"] = Command{Handler: help}
	commands["help"] = Command{Handler: help}

	commands["menu"] = Command{Handler: b.handlerMenu}
	commands["battery"] = Command{Handler: b.handlerBattery}
	commands["bat"] = Command{Handler: b.handlerBattery}
	commands["resa"] = Command{Handler: b.handlerResa}
	commands["logs"] = Command{Handler: b.handlerLogs, Callback: b.callbackLogs}

	codeFSM := b.fsmCodeConversation()
	commands["/code"] = Command{FSM: codeFSM}
	commands[menuCode] = Command{FSM: codeFSM}

	commands["modify"] = Command{Handler: b.handlerModify}

	return commands.start(b)
}
