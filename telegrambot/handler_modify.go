package telegrambot

import (
	"context"
	"fmt"
	"time"

	"github.com/enescakir/emoji"
	"github.com/looplab/fsm"
	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
	"github.com/nmaupu/nuki-logger/model"
	"github.com/rs/zerolog/log"
)

func (bot nukiBot) fsmModifyCommand() *fsm.FSM {
	const (
		metadataPendingModif = "resaPendingModif"
	)

	return fsm.NewFSM(
		"idle",
		fsm.Events{
			{Name: FSMEventDefault, Src: []string{"idle"}, Dst: "wait_resa_id"},
			{Name: "resa_id_received", Src: []string{"idle", "wait_resa_id"}, Dst: "wait_check_in"},
			{Name: "recover_resa_id", Src: []string{"wait_check_in"}, Dst: "wait_resa_id"},
			{Name: "check_in_received", Src: []string{"wait_check_in"}, Dst: "wait_check_out"},
			{Name: "recover_check_in", Src: []string{"wait_check_out"}, Dst: "wait_check_in"},
			{Name: "check_out_received", Src: []string{"wait_check_out"}, Dst: "wait_confirmation"},
			{Name: "recover_check_out", Src: []string{"wait_confirmation"}, Dst: "wait_check_out"},
			{Name: "confirmation_received", Src: []string{"wait_confirmation"}, Dst: "finished"},
			{Name: "reset",
				Src: []string{
					"idle",
					"wait_resa_id",
					"wait_check_in",
					"wait_check_out",
					"wait_confirmation",
					"finished",
				},
				Dst: "idle",
			},
		},
		fsm.Callbacks{
			"wait_resa_id": func(ctx context.Context, e *fsm.Event) {
				log.Debug().Str("callback", "wait_resa_id").Msg("Callback called")
				msg := reinitMetadataMessage(e.FSM)
				msg.ParseMode = telego.ModeMarkdown
				msg.Text = "Enter *reservation* ID"
				waitForUserInput(e.FSM, "resa_id_received")
			},
			"before_resa_id_received": func(ctx context.Context, e *fsm.Event) {
				log.Debug().Str("callback", "before_resa_id_received").Msg("Callback called")
				userInputReceived(e.FSM)
				_ = reinitMetadataMessage(e.FSM)

				data, _ := checkFSMArg(e)

				update, err := getMetadataTelegoUpdate(FSMMetadataTelegoUpdate, e.FSM)
				if err != nil {
					fsmRuntimeErr(e, err.Error(), "reset")
					return
				}

				e.FSM.SetMetadata(metadataPendingModif, &model.ReservationPendingModification{
					ReservationRef: data,
					FromChatID:     update.Message.From.ID,
				})
			},
			"wait_check_in": func(ctx context.Context, e *fsm.Event) {
				log.Debug().Str("callback", "wait_check_in").Msg("Callback called")
				msg := reinitMetadataMessage(e.FSM)
				msg.Text = fmt.Sprintf("Enter check-in time (default: %s)", bot.DefaultCheckIn.Format(model.FormatTimeHoursMinutes))
				waitForUserInput(e.FSM, "check_in_received")
			},
			"before_check_in_received": func(ctx context.Context, e *fsm.Event) {
				log.Debug().Str("callback", "before_check_in_received").Msg("Callback called")
				userInputReceived(e.FSM)
				_ = reinitMetadataMessage(e.FSM)

				data, _ := checkFSMArg(e)

				modif, err := getMetadataReservationPendingModification(metadataPendingModif, e.FSM)
				if err != nil {
					fsmRuntimeErr(e, err.Error(), "reset")
					return
				}

				modif.CheckInTime, err = time.Parse(model.FormatTimeHoursMinutes, data)
				if err != nil {
					fsmRuntimeErr(e, fmt.Sprintf("unable to parse %s", data), "recover_check_in")
					return
				}
			},
			"wait_check_out": func(ctx context.Context, e *fsm.Event) {
				log.Debug().Str("callback", "wait_check_out").Msg("Callback called")
				msg := reinitMetadataMessage(e.FSM)
				msg.Text = fmt.Sprintf("Enter check-out time (default: %s)", bot.DefaultCheckOut.Format(model.FormatTimeHoursMinutes))
				waitForUserInput(e.FSM, "check_out_received")
			},
			"before_check_out_received": func(ctx context.Context, e *fsm.Event) {
				log.Debug().Str("callback", "check_out_received").Msg("Callback called")
				userInputReceived(e.FSM)
				_ = reinitMetadataMessage(e.FSM)

				data, _ := checkFSMArg(e)

				modif, err := getMetadataReservationPendingModification(metadataPendingModif, e.FSM)
				if err != nil {
					fsmRuntimeErr(e, fmt.Sprintf("An error occurred, %s", err.Error()), "reset")
					return
				}

				modif.CheckOutTime, err = time.Parse(model.FormatTimeHoursMinutes, data)
				if err != nil {
					fsmRuntimeErr(e, fmt.Sprintf("unable to parse %s", data), "recover_check_out")
					return
				}
			},
			"wait_confirmation": func(ctx context.Context, e *fsm.Event) {
				log.Debug().Str("callback", "wait_confirmation").Msg("Callback called")
				msg := reinitMetadataMessage(e.FSM)

				modif, err := getMetadataReservationPendingModification(metadataPendingModif, e.FSM)
				if err != nil {
					fsmRuntimeErr(e, fmt.Sprintf("An error occurred, %s", err.Error()), "reset")
					return
				}

				keyboard := tu.InlineKeyboard(tu.InlineKeyboardRow(
					tu.InlineKeyboardButton(fmt.Sprintf("Yes %s", emoji.ThumbsUp.String())).
						WithCallbackData(NewCallbackData("confirmation_received", "yes")),
					tu.InlineKeyboardButton(fmt.Sprintf("No %s", emoji.ThumbsDown.String())).
						WithCallbackData(NewCallbackData("confirmation_received", "no")),
				))

				msg.ReplyMarkup = keyboard
				msg.ParseMode = telego.ModeMarkdown
				m := fmt.Sprintf("%s Do you confirm?\n", emoji.OpenBook.String())
				m += fmt.Sprintf("*Reservation*: %s\n", modif.ReservationRef)
				m += fmt.Sprintf("*Check-in*: %s\n", modif.FormatCheckIn())
				m += fmt.Sprintf("*Check-out*: %s", modif.FormatCheckOut())
				msg.Text = m
			},
			"before_confirmation_received": func(ctx context.Context, e *fsm.Event) {
				log.Debug().Str("callback", "before_confirmation_received").Msg("Callback called")
				msg := reinitMetadataMessage(e.FSM)
				data, _ := checkFSMArg(e)

				modif, err := getMetadataReservationPendingModification(metadataPendingModif, e.FSM)
				if err != nil {
					fsmRuntimeErr(e, fmt.Sprintf("An error occurred, %s", err.Error()), "reset")
					return
				}

				if data == "yes" {
					bot.reservationPendingModificationRoutine.AddPendingModification(*modif)
					msg.Text = "Confirmed!"
				} else {
					msg.Text = "Canceled..."
				}
			},
			"finished": fsmEventFinished,
		},
	)
}
