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
	handlerHelp := func(update telego.Update, msg *telego.SendMessageParams) {
		keys := maps.Keys(commands)
		helpItems := slices.DeleteFunc(keys, func(s string) bool { return !strings.HasPrefix(s, "/") })
		elts := []string{}
		for _, v := range helpItems {
			if v == "/test" {
				continue
			}
			elts = append(elts, fmt.Sprintf("  - %s\t%s", v, commands[v].Description))
		}
		msg.Text = fmt.Sprintf("The following commands are available: \n%s",
			strings.Join(elts, "\n"))
	}

	cmdHelp := Command{Handler: handlerHelp, Description: "Display help"}
	commands["/start"] = cmdHelp
	commands["/help"] = cmdHelp
	commands[menuHelp] = cmdHelp

	commands["/menu"] = Command{Handler: b.handlerMenu, Description: "Show the main menu"}

	cmdBat := Command{Handler: b.handlerBattery, Description: "Display battery details"}
	commands["/battery"] = cmdBat
	commands["/bat"] = cmdBat
	commands[menuBattery] = cmdBat

	cmdResa := Command{Handler: b.handlerResa, Description: "List all reservations"}
	commands["/resa"] = cmdResa
	commands[menuResas] = cmdResa

	logsFSM := b.fsmLogsCommand()
	cmdLogs := Command{StateMachine: logsFSM, Description: "Display Nuki lock logs"}
	commands["/logs"] = cmdLogs
	commands[menuLogs] = cmdLogs

	codeFSM := b.fsmCodeCommand()
	cmdCode := Command{StateMachine: codeFSM, Description: "Display a reservation door code"}
	commands["/code"] = cmdCode
	commands[menuCode] = cmdCode

	commands["/version"] = Command{Handler: b.handlerVersion, Description: "Display bot version"}

	modifyFSM := b.fsmModifyCommand()
	cmdModify := Command{StateMachine: modifyFSM, Description: "Modify check-in/out of a specific reservation"}
	commands["/modify"] = cmdModify
	commands[menuModify] = cmdModify

	cmdListModify := Command{Handler: b.handlerListModify, Description: "List all pending modifications"}
	commands["/listmodify"] = cmdListModify
	commands[menuListModify] = cmdListModify

	commands["/deletemodify"] = Command{StateMachine: b.fsmDeleteModifyCommand(), Description: "Delete a pending modification"}

	commands["/savemodify"] = Command{Handler: b.handlerSavePendingReservationsToCache, Description: "Save all modifications to the cache"}

	commands["/applymodify"] = Command{Handler: b.handlerApplyModify, Description: "Apply all pending modifications now"}

	commands["/test"] = Command{StateMachine: b.fsmTestCommand()}

	b.reservationPendingModificationRoutine.Start(time.Minute * 10)
	return commands.start(b)
}
