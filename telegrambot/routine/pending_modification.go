package routine

import (
	"errors"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/nmaupu/nuki-logger/cache"
	"github.com/nmaupu/nuki-logger/model"
	"github.com/nmaupu/nuki-logger/nukiapi"
	"github.com/rs/zerolog/log"
)

var _ ReservationPendingModificationRoutine = (*reservationPendingModificationRoutine)(nil)

const (
	reservationPendingModificationCacheKey = "reservation-pending-modification"
)

type ErrCannotGetReservationsFromAPI struct {
	error
}

type ReservationPendingModificationRoutine interface {
	Start(checkInterval time.Duration)
	AddPendingModification(r model.ReservationPendingModification)
	GetAllPendingModifications() []model.ReservationPendingModification
	DeletePendingModification(resaRef string)
	AddOnErrorListener(func(rpm *model.ReservationPendingModification, err error))
	AddOnModificationDoneListener(func(rpm *model.ReservationPendingModification))
	SaveToCache() error
	ApplyModificationNow()
}

type reservationPendingModificationRoutine struct {
	cache                       cache.Cache
	pendings                    map[string]*model.ReservationPendingModification
	messages                    chan model.ReservationPendingModification
	mutexRPM                    sync.Mutex
	reservationReader           nukiapi.ReservationsReader
	reservationTimeModifier     nukiapi.ReservationTimeModifier
	mutexListeners              sync.Mutex
	onErrorListeners            []func(rpm *model.ReservationPendingModification, err error)
	onModificationDoneListeners []func(rpm *model.ReservationPendingModification)
	applyNowChan                chan bool
}

func NewReservationPendingModificationRoutine(
	reader nukiapi.ReservationsReader,
	writer nukiapi.ReservationTimeModifier,
	cache cache.Cache) *reservationPendingModificationRoutine {
	return &reservationPendingModificationRoutine{
		cache:                   cache,
		pendings:                make(map[string]*model.ReservationPendingModification),
		messages:                make(chan model.ReservationPendingModification),
		mutexRPM:                sync.Mutex{},
		reservationReader:       reader,
		reservationTimeModifier: writer,
		applyNowChan:            make(chan bool),
	}
}

func (r *reservationPendingModificationRoutine) AddPendingModification(resa model.ReservationPendingModification) {
	resa.LastUpdateTime = time.Now()
	r.messages <- resa
}

func (r *reservationPendingModificationRoutine) DeletePendingModification(resaRef string) {
	r.mutexRPM.Lock()
	defer r.mutexRPM.Unlock()
	delete(r.pendings, resaRef)
	if err := r.SaveToCache(); err != nil {
		log.Error().Err(err).Msg("Unable to save pending modification cache to disk")
	}
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

func (r *reservationPendingModificationRoutine) AddOnErrorListener(f func(rpm *model.ReservationPendingModification, err error)) {
	r.mutexListeners.Lock()
	defer r.mutexListeners.Unlock()
	r.onErrorListeners = append(r.onErrorListeners, f)
}

func (r *reservationPendingModificationRoutine) AddOnModificationDoneListener(f func(rpm *model.ReservationPendingModification)) {
	r.mutexListeners.Lock()
	defer r.mutexListeners.Unlock()
	r.onModificationDoneListeners = append(r.onModificationDoneListeners, f)
}

func (r *reservationPendingModificationRoutine) ApplyModificationNow() {
	r.applyNowChan <- true
}

func (r *reservationPendingModificationRoutine) Start(checkInterval time.Duration) {
	go func() {
		interrupt := make(chan os.Signal, 1)
		signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
		ticker := time.NewTicker(checkInterval)
		defer ticker.Stop()
		tickerTimeoutDoneChecker := time.NewTicker(time.Hour * 2)
		defer tickerTimeoutDoneChecker.Stop()

		if err := r.loadFromCache(); err != nil {
			log.Error().Err(err).Msg("Unable to load pending reservations cache from disk")
		}

		for {
			select {
			case <-tickerTimeoutDoneChecker.C: // Check for done pending reservations to delete
				r.mutexRPM.Lock()
				for resaRef, pending := range r.pendings {
					if pending.ModificationDone &&
						!pending.LastUpdateTime.IsZero() &&
						time.Since(pending.LastUpdateTime) > model.TimeoutReservationPendingModificationDoneDuration {
						delete(r.pendings, resaRef)
					}
				}
				if err := r.SaveToCache(); err != nil {
					log.Error().Err(err).Msg("Unable to save pending modification cache to disk")
				}
				r.mutexRPM.Unlock()
			case <-ticker.C: // Check for all pending modifications not done yet
				r.apply()
			case <-r.applyNowChan: // Apply all pending modifications now
				r.apply()
			case msg := <-r.messages: // New pending modification
				log.Debug().
					Object("pending_resa", msg).
					Msg("Adding resa to pending list")
				r.mutexRPM.Lock()
				r.pendings[msg.ReservationRef] = &msg
				if err := r.SaveToCache(); err != nil {
					log.Error().Err(err).Msg("Unable to save pending modification cache to disk")
					r.dispatchErrorsToListeners(&msg, err)
				}
				r.mutexRPM.Unlock()
			case <-interrupt:
				log.Info().Msg("Stopping reservation pending modification routine")
				return
			}
		}
	}()
}

func (r *reservationPendingModificationRoutine) apply() {
	log.Info().Msg("Processing pending reservations")

	// Get all reservations from API
	allResas, err := r.reservationReader.Execute()
	if err != nil {
		r.dispatchErrorsToListeners(nil, ErrCannotGetReservationsFromAPI{err})
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
				r.dispatchErrorsToListeners(pendingResa, err)
				continue
			}
			pendingResa.ModificationDone = true
			pendingResa.LastUpdateTime = time.Now()
			linkedResa := resa
			pendingResa.LinkedReservation = &linkedResa
			r.dispatchModificationDoneToListeners(pendingResa)
		}
	}
	r.mutexRPM.Unlock()
}

func (r *reservationPendingModificationRoutine) dispatchErrorsToListeners(rpm *model.ReservationPendingModification, e error) {
	r.mutexListeners.Lock()
	defer r.mutexListeners.Unlock()
	for _, fn := range r.onErrorListeners {
		if fn != nil {
			fn(rpm, e)
		}
	}
}

func (r *reservationPendingModificationRoutine) dispatchModificationDoneToListeners(rpm *model.ReservationPendingModification) {
	r.mutexListeners.Lock()
	defer r.mutexListeners.Unlock()
	for _, fn := range r.onModificationDoneListeners {
		if fn != nil {
			fn(rpm)
		}
	}
}

func (r *reservationPendingModificationRoutine) SaveToCache() error {
	if r.cache == nil {
		return cache.ErrCacheNoClient
	}
	return r.cache.Save(reservationPendingModificationCacheKey, r.pendings)
}

func (r *reservationPendingModificationRoutine) loadFromCache() error {
	if r.cache == nil {
		return cache.ErrCacheNoClient
	}

	if r.pendings == nil {
		r.pendings = make(map[string]*model.ReservationPendingModification)
	}

	err := r.cache.Load(reservationPendingModificationCacheKey, &r.pendings)
	switch {
	case errors.Is(err, memcache.ErrCacheMiss):
		return nil
	case errors.Is(err, memcache.ErrNoServers):
		return nil
	default:
		return err
	}
}
