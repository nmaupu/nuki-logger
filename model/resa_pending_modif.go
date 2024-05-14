package model

import (
	"time"

	"github.com/rs/zerolog"
)

var _ zerolog.LogObjectMarshaler = (*ReservationPendingModification)(nil)

const (
	FormatTimeHoursMinutes                            = "15:04"
	TimeoutReservationPendingModificationDoneDuration = time.Hour * 72
)

// reservationPendingModification registers a pending modification. This will be applied when the reservation appears on Nuki side
// Check in/out time is an int32 representing the number of minutes from midnight stored as a go time
type ReservationPendingModification struct {
	ReservationRef    string                   `json:"reservation_ref"`
	CheckInTime       time.Time                `json:"check_in_time"`
	CheckOutTime      time.Time                `json:"check_out_time"`
	ModificationDone  bool                     `json:"modification_done"`
	LinkedReservation *NukiReservationResponse `json:"linked_reservation"`
	FromChatID        int64                    `json:"from_chat_id"`
	LastUpdateTime    time.Time                `json:"last_update_time"`
}

func (r ReservationPendingModification) FormatCheckIn() string {
	return r.CheckInTime.Format(FormatTimeHoursMinutes)
}

func (r ReservationPendingModification) FormatCheckOut() string {
	return r.CheckOutTime.Format(FormatTimeHoursMinutes)
}

func (r ReservationPendingModification) MarshalZerologObject(e *zerolog.Event) {
	e.Str("reservation_ref", r.ReservationRef).
		Time("check_in", r.CheckInTime).
		Time("check_out", r.CheckOutTime).
		Bool("modification_done", r.ModificationDone).
		Bool("linked_reservation_is_nil", r.LinkedReservation == nil).
		Int64("from_chat_id", r.FromChatID)
}

func MinutesFromMidnight(t time.Time) int32 {
	h, m, _ := t.Clock()
	return int32(h*60 + m)
}
