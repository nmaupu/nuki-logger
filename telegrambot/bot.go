package telegrambot

import (
	"fmt"
	"slices"
	"strings"

	"github.com/mymmrac/telego"
	"github.com/nmaupu/nuki-logger/messaging"
	"github.com/nmaupu/nuki-logger/nukiapi"
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
		keys := maps.Keys(commands)
		helpItems := slices.DeleteFunc(keys, func(s string) bool { return !strings.HasPrefix(s, "/") })
		msg.Text = fmt.Sprintf("The following commands are available: %s", strings.Join(helpItems, ", "))
	}
	commands["/start"] = Command{Handler: help}
	commands["/help"] = Command{Handler: help}
	commands[menuHelp] = Command{Handler: help}

	commands["/battery"] = Command{Handler: b.handlerBattery}
	commands["/bat"] = Command{Handler: b.handlerBattery}
	commands[menuBattery] = Command{Handler: b.handlerBattery}

	commands["/resa"] = Command{Handler: b.handlerResa}
	commands[menuResa] = Command{Handler: b.handlerResa}

	logsFSM := b.fsmLogsCommand()
	commands["/logs"] = Command{StateMachine: FSM{logsFSM}}
	commands[menuLogs] = Command{StateMachine: FSM{logsFSM}}

	commands["/menu"] = Command{Handler: b.handlerMenu}

	codeFSM := b.fsmCodeCommand()
	commands["/code"] = Command{StateMachine: FSM{codeFSM}}
	commands[menuCode] = Command{StateMachine: FSM{codeFSM}}

	commands["/name"] = Command{StateMachine: FSM{b.fsmNameCommand()}}

	commands["/modify"] = Command{Handler: b.handlerModify}
	commands[menuModify] = Command{Handler: b.handlerModify}

	return commands.start(b)
}
