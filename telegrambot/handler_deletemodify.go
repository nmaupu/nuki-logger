package telegrambot

import (
	"context"
	"fmt"

	"github.com/looplab/fsm"
	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
	"github.com/rs/zerolog/log"
)

func (bot nukiBot) fsmDeleteModifyCommand() *fsm.FSM {
	return fsm.NewFSM(
		"idle",
		fsm.Events{
			{Name: FSMEventDefault, Src: []string{"idle"}, Dst: "waiting_for_resa"},
			{Name: "resa_received", Src: []string{"idle", "waiting_for_resa"}, Dst: "finished"},
			{Name: "reset", Src: []string{"idle", "waiting_for_resa", "finished"}, Dst: "idle"},
		},
		fsm.Callbacks{
			FSMEventDefault:        bot.fsmEventDeleteModifyDefault,
			"before_resa_received": bot.fsmEventDeleteModifyReceived,
			"finished":             fsmEventFinished,
		},
	)
}

func (bot nukiBot) fsmEventDeleteModifyDefault(ctx context.Context, e *fsm.Event) {
	log.Debug().Str("callback", FSMEventDefault).Msg("Callback called")
	msg := reinitMetadataMessage(e.FSM)

	modifs := bot.reservationPendingModificationRoutine.GetAllPendingModifications()
	if len(modifs) == 0 {
		fsmRuntimeErr(e, "No pending modification available", "reset")
		return

	}

	var keyboardButtons []telego.InlineKeyboardButton
	for _, modif := range modifs {
		keyboardButtons = append(keyboardButtons,
			tu.InlineKeyboardButton(fmt.Sprintf("%s (%s - %s)", modif.ReservationRef, modif.FormatCheckIn(), modif.FormatCheckOut())).
				WithCallbackData(NewCallbackData("resa_received", modif.ReservationRef)))
	}

	msg.ReplyMarkup = tu.InlineKeyboard(keyboardButtons)
	msg.ParseMode = telego.ModeMarkdown
	msg.Text = "What *pending modification* do you want to remove?"
	msg.ProtectContent = true

}

func (bot *nukiBot) fsmEventDeleteModifyReceived(ctx context.Context, e *fsm.Event) {
	log.Debug().Str("callback", "resa_received").Msg("Callback called")
	msg := reinitMetadataMessage(e.FSM)

	data, err := checkFSMArg(e)
	if err != nil {
		msg.Text = fmt.Sprintf("An error occurred: %s", err)
		return
	}

	bot.reservationPendingModificationRoutine.DeletePendingModification(data)
	msg.Text = "Done\nClick on /listmodify to display the new list"
}
