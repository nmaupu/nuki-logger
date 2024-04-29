package nukiapi

const (
	Api                  = "https://api.nuki.io"
	LogsEndpoint         = "smartlock/%d/log"
	ReservationsEndpoint = "address/%d/reservation"
)

type APICaller struct {
	Token string
}
