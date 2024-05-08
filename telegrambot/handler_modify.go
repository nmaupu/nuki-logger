package telegrambot

import (
	"time"

	"github.com/looplab/fsm"
	"github.com/mymmrac/telego" // tu "github.com/mymmrac/telego/telegoutil"
)

// reservationPendingModification registers a pending modification. This will be applied when the reservation appears on Nuki side
// Check in/out time is an int32 representing the number of minutes from midnight stored as a go time
type ReservationPendingModification struct {
	ReservationID string    `json:"reservation_id"`
	CheckInTime   time.Time `json:"check_in_time"`
	CheckOutTime  time.Time `json:"check_out_time"`
	FSM           *fsm.FSM
}

func NewReservationPendingModification() ReservationPendingModification {
	r := ReservationPendingModification{}
	r.FSM = fsm.NewFSM(
		"init",
		fsm.Events{
			{Name: "enter-resa-id", Src: []string{"init"}, Dst: "resa-id-entered"},
			{Name: "enter-check-in", Src: []string{"resa-id-entered"}, Dst: "check-in-entered"},
			{Name: "enter-check-out", Src: []string{"check-in-entered"}, Dst: "check-out-entered"},
			{Name: "confirm", Src: []string{"check-out-entered"}, Dst: "confirmed"},
		},
		fsm.Callbacks{},
	)
	return r
}

func minutesFromMidnight(t time.Time) int32 {
	h, m, _ := t.Clock()
	return int32(h*60 + m)
}

func (b *nukiBot) handlerModify(update telego.Update, msg *telego.SendMessageParams) {
	msg.Text = "Not implemented yet."
}
