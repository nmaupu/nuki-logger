package nukiapi

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/rs/zerolog/log"
)

const (
	Api                   = "https://api.nuki.io"
	LogsEndpoint          = "smartlock/%d/log"
	ReservationsEndpoint  = "address/%d/reservation"
	SmartlockEndpoint     = "smartlock/%d"
	SmartlockAuthEndpoint = "smartlock/%d/auth"
)

type APICaller struct {
	Token string
}

func (c APICaller) execAPIGet(requestURL string) ([]byte, error) {
	log.Debug().
		Str("request_url", requestURL).
		Msg("Calling Nuki API (GET)")
	httpReq, err := http.NewRequest(http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, err
	}
	httpReq.Header = http.Header{
		"Content-Type":  {"application/json"},
		"Authorization": {fmt.Sprintf("Bearer %s", c.Token)},
	}
	client := http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("error while querying Nuki API (status: %s): %s", resp.Status, string(body))
	}

	return body, nil
}

func (c APICaller) execAPIPost(requestURL string, body []byte) ([]byte, error) {
	log.Debug().
		Str("request_url", requestURL).
		Msg("Calling Nuki API (POST)")
	httpReq, err := http.NewRequest(http.MethodPost, requestURL, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header = http.Header{
		"Content-Type":  {"application/json"},
		"Authorization": {fmt.Sprintf("Bearer %s", c.Token)},
	}
	client := http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}

	bodyRes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("error while querying Nuki API (status: %s): %s", resp.Status, string(bodyRes))
	}

	return bodyRes, nil
}
