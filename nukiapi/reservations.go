package nukiapi

import (
	"encoding/json"
	"fmt"
	"github.com/nmaupu/nuki-logger/model"
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
