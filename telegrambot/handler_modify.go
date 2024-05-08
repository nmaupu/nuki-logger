package telegrambot

import (
	"context"
	"fmt"
	"time"

	"github.com/enescakir/emoji"
	"github.com/looplab/fsm"
	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
	"github.com/rs/zerolog/log"
)

const (
	FormatTimeHoursMinutes = "15:04"
)

var reservationPendingModifications = map[string]ReservationPendingModification{}

// reservationPendingModification registers a pending modification. This will be applied when the reservation appears on Nuki side
// Check in/out time is an int32 representing the number of minutes from midnight stored as a go time
type ReservationPendingModification struct {
	ReservationID string    `json:"reservation_id"`
	CheckInTime   time.Time `json:"check_in_time"`
	CheckOutTime  time.Time `json:"check_out_time"`
}

func (r ReservationPendingModification) FormatCheckIn() string {
	return r.CheckInTime.Format(FormatTimeHoursMinutes)
}

func (r ReservationPendingModification) FormatCheckOut() string {
	return r.CheckOutTime.Format(FormatTimeHoursMinutes)
}

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

				e.FSM.SetMetadata(metadataPendingModif, &ReservationPendingModification{
					ReservationID: data,
				})
			},
			"wait_check_in": func(ctx context.Context, e *fsm.Event) {
				log.Debug().Str("callback", "wait_check_in").Msg("Callback called")
				msg := reinitMetadataMessage(e.FSM)
				msg.Text = fmt.Sprintf("Enter check-in time (format: %s)", FormatTimeHoursMinutes)
				waitForUserInput(e.FSM, "check_in_received")
			},
			"before_check_in_received": func(ctx context.Context, e *fsm.Event) {
				log.Debug().Str("callback", "before_check_in_received").Msg("Callback called")
				userInputReceived(e.FSM)
				_ = reinitMetadataMessage(e.FSM)

				data, _ := checkFSMArg(e)

				modif, err := getMetadataReservationPendingModification(metadataPendingModif, e.FSM)
				if err != nil {
					fsmRuntimeErr(e, err, "reset")
					return
				}

				modif.CheckInTime, err = time.Parse(FormatTimeHoursMinutes, data)
				if err != nil {
					fsmRuntimeErr(e, fmt.Errorf("unable to parse %s", data), "recover_check_in")
				}
			},
			"wait_check_out": func(ctx context.Context, e *fsm.Event) {
				log.Debug().Str("callback", "wait_check_out").Msg("Callback called")
				msg := reinitMetadataMessage(e.FSM)
				msg.Text = fmt.Sprintf("Enter check-out time (format: %s)", FormatTimeHoursMinutes)
				waitForUserInput(e.FSM, "check_out_received")
			},
			"before_check_out_received": func(ctx context.Context, e *fsm.Event) {
				log.Debug().Str("callback", "check_out_received").Msg("Callback called")
				userInputReceived(e.FSM)
				_ = reinitMetadataMessage(e.FSM)

				data, _ := checkFSMArg(e)

				modif, err := getMetadataReservationPendingModification(metadataPendingModif, e.FSM)
				if err != nil {
					fsmRuntimeErr(e, fmt.Errorf("An error occurred, %s", err.Error()), "reset")
					return
				}

				modif.CheckOutTime, err = time.Parse(FormatTimeHoursMinutes, data)
				if err != nil {
					fsmRuntimeErr(e, fmt.Errorf("unable to parse %s", data), "recover_check_out")
				}
			},
			"wait_confirmation": func(ctx context.Context, e *fsm.Event) {
				log.Debug().Str("callback", "wait_confirmation").Msg("Callback called")
				msg := reinitMetadataMessage(e.FSM)

				modif, err := getMetadataReservationPendingModification(metadataPendingModif, e.FSM)
				if err != nil {
					fsmRuntimeErr(e, fmt.Errorf("An error occurred, %s", err.Error()), "reset")
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
				m += fmt.Sprintf("*Reservation*: %s\n", modif.ReservationID)
				m += fmt.Sprintf("*Check-in*: %s\n", modif.FormatCheckIn())
				m += fmt.Sprintf("*Check-out*: %s", modif.FormatCheckOut())
				msg.Text = m
			},
			"before_confirmation_received": func(ctx context.Context, e *fsm.Event) {
				log.Debug().Str("callback", "before_confirmation_received").Msg("Callback called")
				msg := reinitMetadataMessage(e.FSM)
				data, _ := checkFSMArg(e)

				if data == "yes" {
					msg.Text = "Confirmed!"
				} else {
					msg.Text = "Canceled..."
				}
			},
			"finished": fsmEventFinished,
		},
	)
}

func minutesFromMidnight(t time.Time) int32 {
	h, m, _ := t.Clock()
	return int32(h*60 + m)
}
