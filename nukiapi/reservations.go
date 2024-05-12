package nukiapi

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/nmaupu/nuki-logger/model"
)

var (
	reservationsCache           = map[string]string{}
	reservationsCacheLastUpdate time.Time
	reservationsCacheTimeout    = time.Hour * 2
)

type ReservationsReader struct {
	APICaller
	AddressID int64
}

func (r ReservationsReader) Execute() ([]model.NukiReservationResponse, error) {
	if r.AddressID == 0 {
		return nil, fmt.Errorf("addressid is mandatory")
	}
	if r.Token == "" {
		return nil, fmt.Errorf("token is mandatory")
	}

	requestURL := fmt.Sprintf("%s/%s", Api, fmt.Sprintf(ReservationsEndpoint, r.AddressID))

	body, err := r.execAPIGet(requestURL)
	if err != nil {
		return nil, err
	}

	var responses []model.NukiReservationResponse
	err = json.Unmarshal(body, &responses)
	return responses, err
}

// getReservationName returns the name associated to a reservation
func (r ReservationsReader) GetReservationName(ref string) (string, error) {
	if ref == "" {
		return "", nil
	}

	var ok bool
	var reservationName string
	reservationName, ok = reservationsCache[ref]
	if ok &&
		!reservationsCacheLastUpdate.IsZero() &&
		time.Since(reservationsCacheLastUpdate) < reservationsCacheTimeout {
		return reservationName, nil
	}

	// getting real person's name from address API
	reservations, err := r.Execute()
	if err != nil {
		return "", err
	}
	for _, resa := range reservations {
		reservationsCache[resa.Reference] = resa.Name
	}
	reservationsCacheLastUpdate = time.Now()

	reservationName, ok = reservationsCache[ref]
	if !ok {
		return "", fmt.Errorf("unable to find ref '%s'", ref)
	}
	return reservationName, nil
}

type ReservationTimeModifier struct {
	APICaller
	AddressID int64
}

func (r ReservationTimeModifier) Execute(resaID string, checkIn int32, checkOut int32) error {
	if r.AddressID == 0 {
		return fmt.Errorf("addressid is mandatory")
	}
	if r.Token == "" {
		return fmt.Errorf("token is mandatory")
	}

	requestURL := fmt.Sprintf("%s/%s/%s/update/accesstimes", Api, fmt.Sprintf(ReservationsEndpoint, r.AddressID), resaID)
	bodyPost := struct {
		CheckInTime  int32 `json:"checkInTime"`
		CheckOutTime int32 `json:"checkOutTime"`
	}{
		CheckInTime:  checkIn,
		CheckOutTime: checkOut,
	}
	bodyJSON, err := json.Marshal(bodyPost)
	if err != nil {
		return err
	}

	body, err := r.execAPIPost(requestURL, bodyJSON)
	if err != nil {
		return fmt.Errorf("Unable to send request, err=%v, body=%s", err, string(body))
	}

	return nil
}
