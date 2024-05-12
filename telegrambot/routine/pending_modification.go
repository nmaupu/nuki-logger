package routine

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/nmaupu/nuki-logger/model"
	"github.com/nmaupu/nuki-logger/nukiapi"
	"github.com/rs/zerolog/log"
)

var _ ReservationPendingModificationRoutine = (*reservationPendingModificationRoutine)(nil)

type ErrCannotGetReservationsFromAPI struct {
	error
}

type ReservationPendingModificationRoutine interface {
	Start(checkInterval time.Duration)
	AddPendingModification(r model.ReservationPendingModification)
	GetAllPendingModifications() []model.ReservationPendingModification
	DeletePendingModification(resaID string)
}

type reservationPendingModificationRoutine struct {
	pendings                map[string]*model.ReservationPendingModification
	messages                chan model.ReservationPendingModification
	mutexRPM                sync.Mutex
	reservationReader       nukiapi.ReservationsReader
	reservationTimeModifier nukiapi.ReservationTimeModifier
	errorsCallback          func(err error)
}

func NewReservationPendingModificationRoutine(reader nukiapi.ReservationsReader, writer nukiapi.ReservationTimeModifier, errorsCallback func(err error)) *reservationPendingModificationRoutine {
	return &reservationPendingModificationRoutine{
		pendings:                make(map[string]*model.ReservationPendingModification),
		messages:                make(chan model.ReservationPendingModification),
		mutexRPM:                sync.Mutex{},
		reservationReader:       reader,
		reservationTimeModifier: writer,
		errorsCallback:          errorsCallback,
	}
}

func (r *reservationPendingModificationRoutine) AddPendingModification(resa model.ReservationPendingModification) {
	r.messages <- resa
}

func (r *reservationPendingModificationRoutine) DeletePendingModification(resaID string) {
	r.mutexRPM.Lock()
	defer r.mutexRPM.Unlock()
	delete(r.pendings, resaID)
}

func (r *reservationPendingModificationRoutine) GetAllPendingModifications() []model.ReservationPendingModification {
	res := []model.ReservationPendingModification{}
	r.mutexRPM.Lock()
	defer r.mutexRPM.Unlock()
	for _, v := range r.pendings {
		res = append(res, *v)
	}
	return res
}

func (r *reservationPendingModificationRoutine) Start(checkInterval time.Duration) {
	go func() {
		interrupt := make(chan os.Signal, 1)
		signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
		ticker := time.NewTicker(checkInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C: // Check for all pending modifications not done yet
				log.Info().Msg("Processing pending reservations")

				// Get all reservations from API
				allResas, err := r.reservationReader.Execute()
				if err != nil {
					log.Error().Err(err).Send()
					if r.errorsCallback != nil {
						r.errorsCallback(ErrCannotGetReservationsFromAPI{err})
					}
				}

				// Getting all pending modifications and do the change
				r.mutexRPM.Lock()
				for _, resa := range allResas {
					pendingResa, ok := r.pendings[resa.Reference]
					if ok && !pendingResa.ModificationDone { // found one to process
						log.Info().
							Object("pending_resa", pendingResa).
							Msg("Processing pending resa change")
						err := r.reservationTimeModifier.Execute(
							resa.ID,
							model.MinutesFromMidnight(pendingResa.CheckInTime),
							model.MinutesFromMidnight(pendingResa.CheckOutTime),
						)
						if err != nil {
							log.Error().Err(err).Msg("Unable to process pending modification")
							continue
						}
						pendingResa.ModificationDone = true
						linkedResa := resa
						pendingResa.LinkedReservation = &linkedResa
					}
				}
				r.mutexRPM.Unlock()
			case msg := <-r.messages: // New pending modification
				log.Debug().
					Object("pending_resa", msg).
					Msg("Adding resa to pending list")
				r.mutexRPM.Lock()
				r.pendings[msg.ReservationID] = &msg
				r.mutexRPM.Unlock()
			case <-interrupt:
				log.Info().Msg("Stopping reservation pending modification routine")
				return
			}
		}
	}()
}
