package cli

import (
	"fmt"
	"github.com/nmaupu/nuki-logger/nukiapi"
	"time"
)

var (
	reservationsCache           = map[string]string{}
	reservationsCacheLastUpdate time.Time
	reservationsCacheTimeout    = time.Hour * 2
)

// getReservationName returns the name associated to a reservation
func getReservationName(ref string, config *Config) (string, error) {
	if ref == "" {
		return "", nil
	}

	var ok bool
	var reservationName string
	reservationName, ok = reservationsCache[ref]
	if ok &&
		!reservationsCacheLastUpdate.IsZero() &&
		time.Now().Sub(reservationsCacheLastUpdate) < reservationsCacheTimeout {
		return reservationName, nil
	}

	// getting real person's name from address API
	reservationsReader := nukiapi.ReservationsReader{
		APICaller: nukiapi.APICaller{Token: config.NukiAPIToken},
		AddressID: config.AddressID,
	}
	reservations, err := reservationsReader.Execute()
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
