package telegrambot

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/mymmrac/telego"
	"github.com/nmaupu/nuki-logger/messaging"
	"github.com/nmaupu/nuki-logger/nukiapi"
	tgbroutine "github.com/nmaupu/nuki-logger/telegrambot/routine"
	"github.com/rs/zerolog/log"
	"golang.org/x/exp/maps"
)

type NukiBot interface {
	Start() error
	AddFilter(FilterFunc)
}

type nukiBot struct {
	Sender                                *messaging.TelegramSender
	LogsReader                            nukiapi.LogsReader
	SmartlockReader                       nukiapi.SmartlockReader
	ReservationsReader                    nukiapi.ReservationsReader
	SmartlockAuthReader                   nukiapi.SmartlockAuthReader
	filters                               []FilterFunc
	reservationPendingModificationRoutine tgbroutine.ReservationPendingModificationRoutine
}

func NewNukiBot(sender *messaging.TelegramSender,
	logsReader nukiapi.LogsReader,
	smartlockReader nukiapi.SmartlockReader,
	reservationsReader nukiapi.ReservationsReader,
	smartlockAuthReader nukiapi.SmartlockAuthReader,
	filters ...FilterFunc) NukiBot {

	resaTimeModifier := nukiapi.ReservationTimeModifier{
		APICaller: nukiapi.APICaller{Token: reservationsReader.Token},
		AddressID: reservationsReader.AddressID,
	}
	return &nukiBot{
		Sender:              sender,
		LogsReader:          logsReader,
		SmartlockReader:     smartlockReader,
		ReservationsReader:  reservationsReader,
		SmartlockAuthReader: smartlockAuthReader,
		filters:             filters,
		reservationPendingModificationRoutine: tgbroutine.NewReservationPendingModificationRoutine(
			reservationsReader,
			resaTimeModifier,
			func(e error) {
				log.Error().Err(e).Msg("An error occurred processing pending reservations")
			},
		),
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

	modifyFSM := b.fsmModifyCommand()
	commands["/modify"] = Command{StateMachine: modifyFSM}
	commands[menuModify] = Command{StateMachine: modifyFSM}

	commands["/listmodify"] = Command{Handler: b.handlerListModify}
	commands[menuListModify] = Command{Handler: b.handlerListModify}

	commands["/deletemodify"] = Command{StateMachine: b.fsmDeleteModifyCommand()}

	b.reservationPendingModificationRoutine.Start(time.Second * 10)
	return commands.start(b)
}
