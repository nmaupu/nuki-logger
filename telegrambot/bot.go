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
	AddFilter(FilterFunc)
}

type nukiBot struct {
	sender              *messaging.TelegramSender
	logsReader          nukiapi.LogsReader
	smartlockReader     nukiapi.SmartlockReader
	reservationsReader  nukiapi.ReservationsReader
	smartlockAuthReader nukiapi.SmartlockAuthReader
	filters             []FilterFunc
}

func NewNukiBot(sender *messaging.TelegramSender,
	logsReader nukiapi.LogsReader,
	smartlockReader nukiapi.SmartlockReader,
	reservationsReader nukiapi.ReservationsReader,
	smartlockAuthReader nukiapi.SmartlockAuthReader,
	filters ...FilterFunc) NukiBot {

	return &nukiBot{
		sender:              sender,
		logsReader:          logsReader,
		smartlockReader:     smartlockReader,
		reservationsReader:  reservationsReader,
		smartlockAuthReader: smartlockAuthReader,
		filters:             filters,
	}
}

func (b *nukiBot) AddFilter(f FilterFunc) {
	b.filters = append(b.filters, f)
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
	commands["code"] = Command{Handler: b.handlerCode, Callback: b.callbackCode}

	return commands.start(b)
}
