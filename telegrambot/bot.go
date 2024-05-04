package telegrambot

import (
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/nmaupu/nuki-logger/messaging"
	"github.com/nmaupu/nuki-logger/nukiapi"
	sf "github.com/sa-/slicefunk"
	"golang.org/x/exp/maps"
)

type NukiBot interface {
	Start() error
}

type nukiBot struct {
	sender              *messaging.TelegramSender
	logsReader          nukiapi.LogsReader
	smartlockReader     nukiapi.SmartlockReader
	reservationsReader  nukiapi.ReservationsReader
	smartlockAuthReader nukiapi.SmartlockAuthReader
}

func NewNukiBot(sender *messaging.TelegramSender,
	logsReader nukiapi.LogsReader,
	smartlockReader nukiapi.SmartlockReader,
	reservationsReader nukiapi.ReservationsReader,
	smartlockAuthReader nukiapi.SmartlockAuthReader) NukiBot {

	return &nukiBot{
		sender:              sender,
		logsReader:          logsReader,
		smartlockReader:     smartlockReader,
		reservationsReader:  reservationsReader,
		smartlockAuthReader: smartlockAuthReader,
	}
}

func (b *nukiBot) Start() error {
	commands := Commands{}
	help := func(update tgbotapi.Update, msg *tgbotapi.MessageConfig) {
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
