package telegrambot

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
	"github.com/nmaupu/nuki-logger/cache"
	"github.com/nmaupu/nuki-logger/messaging"
	"github.com/nmaupu/nuki-logger/model"
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
	bot                                   *telego.Bot
	Sender                                *messaging.TelegramSender
	LogsReader                            nukiapi.LogsReader
	SmartlockReader                       nukiapi.SmartlockReader
	ReservationsReader                    nukiapi.ReservationsReader
	SmartlockAuthReader                   nukiapi.SmartlockAuthReader
	filters                               []FilterFunc
	reservationPendingModificationRoutine tgbroutine.ReservationPendingModificationRoutine
	DefaultCheckIn                        time.Time
	DefaultCheckOut                       time.Time
}

func NewNukiBot(sender *messaging.TelegramSender,
	logsReader nukiapi.LogsReader,
	smartlockReader nukiapi.SmartlockReader,
	reservationsReader nukiapi.ReservationsReader,
	smartlockAuthReader nukiapi.SmartlockAuthReader,
	defaultCheckIn time.Time,
	defaultCheckOut time.Time,
	cache cache.Cache,
	filters ...FilterFunc) (NukiBot, error) {

	bot, err := telego.NewBot(sender.Token)
	if err != nil {
		return nil, err
	}
	resaTimeModifier := nukiapi.ReservationTimeModifier{
		APICaller: nukiapi.APICaller{Token: reservationsReader.Token},
		AddressID: reservationsReader.AddressID,
	}
	resaPendingModifRoutine := tgbroutine.NewReservationPendingModificationRoutine(reservationsReader, resaTimeModifier, cache)
	return &nukiBot{
		bot:                                   bot,
		Sender:                                sender,
		LogsReader:                            logsReader,
		SmartlockReader:                       smartlockReader,
		ReservationsReader:                    reservationsReader,
		SmartlockAuthReader:                   smartlockAuthReader,
		filters:                               filters,
		reservationPendingModificationRoutine: resaPendingModifRoutine,
		DefaultCheckIn:                        defaultCheckIn,
		DefaultCheckOut:                       defaultCheckOut,
	}, nil
}

func (b *nukiBot) AddFilter(f FilterFunc) {
	b.filters = append(b.filters, f)
}

func (b *nukiBot) Start() error {
	b.reservationPendingModificationRoutine.AddOnErrorListener(func(rpm *model.ReservationPendingModification, e error) {
		log.Error().Err(e).Msg("An error occurred processing pending modifications")
		if rpm != nil {
			_, _ = b.bot.SendMessage(tu.Message(tu.ID(rpm.FromChatID), fmt.Sprintf("An error occurred processing pending modification, err=%v", e)))
		}
	})

	b.reservationPendingModificationRoutine.AddOnModificationDoneListener(func(rpm *model.ReservationPendingModification) {
		if rpm == nil {
			log.Warn().Msgf("onModificationListener callback called with a nil modification")
			return
		}

		log.Debug().
			Str("ref", rpm.ReservationRef).
			Str("check_in", rpm.FormatCheckIn()).
			Str("check_out", rpm.FormatCheckOut()).
			Msg("Pending modification done")
		_, _ = b.bot.SendMessage(tu.Message(
			tu.ID(rpm.FromChatID),
			fmt.Sprintf("Pending modification done for %s (%s -> %s)", rpm.ReservationRef, rpm.FormatCheckIn(), rpm.FormatCheckOut())),
		)
	})

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

	b.reservationPendingModificationRoutine.Start(time.Minute * 10)
	return commands.start(b)
}
