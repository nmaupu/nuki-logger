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
	AddFilter(FilterFunc)
}

type nukiBot struct {
	Sender              *messaging.TelegramSender
	LogsReader          nukiapi.LogsReader
	SmartlockReader     nukiapi.SmartlockReader
	ReservationsReader  nukiapi.ReservationsReader
	SmartlockAuthReader nukiapi.SmartlockAuthReader
	filters             []FilterFunc
}

func NewNukiBot(sender *messaging.TelegramSender,
	logsReader nukiapi.LogsReader,
	smartlockReader nukiapi.SmartlockReader,
	reservationsReader nukiapi.ReservationsReader,
	smartlockAuthReader nukiapi.SmartlockAuthReader,
	filters ...FilterFunc) NukiBot {

	return &nukiBot{
		Sender:              sender,
		LogsReader:          logsReader,
		SmartlockReader:     smartlockReader,
		ReservationsReader:  reservationsReader,
		SmartlockAuthReader: smartlockAuthReader,
		filters:             filters,
	}
}

func (b *nukiBot) AddFilter(f FilterFunc) {
	b.filters = append(b.filters, f)
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

	commands["/menu"] = Command{Handler: b.handlerMenu}

	commands["/battery"] = Command{Handler: b.handlerBattery}
	commands["/bat"] = Command{Handler: b.handlerBattery}
	commands[menuBattery] = Command{Handler: b.handlerBattery}

	commands["/resa"] = Command{Handler: b.handlerResa}
	commands[menuResas] = Command{Handler: b.handlerResa}

	logsFSM := b.fsmLogsCommand()
	commands["/logs"] = Command{StateMachine: logsFSM}
	commands[menuLogs] = Command{StateMachine: logsFSM}

	codeFSM := b.fsmCodeCommand()
	commands["/code"] = Command{StateMachine: codeFSM}
	commands[menuCode] = Command{StateMachine: codeFSM}

	commands["/name"] = Command{StateMachine: b.fsmNameCommand()}

	commands["/modify"] = Command{Handler: b.handlerModify}
	commands[menuModify] = Command{Handler: b.handlerModify}

	return commands.start(b)
}
