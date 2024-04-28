package nukiapi

import (
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
	"nuki-logger/model"
	"strings"
	"time"
)

const (
	NukiApi         = "https://api.nuki.io"
	NukiLogEndpoint = "smartlock/%s/log"
)

type LogsReader struct {
	SmartlockID string
	Token       string
	Limit       int
	FromDate    time.Time
	ToDate      time.Time
}

func (r LogsReader) Execute() ([]model.NukiSmartlockLogResponse, error) {
	if r.SmartlockID == "" {
		return nil, fmt.Errorf("smartlockid is mandatory")
	}
	if r.Token == "" {
		return nil, fmt.Errorf("token is mandatory")
	}
	if r.Limit < 0 {
		r.Limit = 1
	}
	if r.Limit > 50 {
		r.Limit = 50
	}
	getParams := []string{fmt.Sprintf("limit=%d", r.Limit)}
	if !r.FromDate.IsZero() {
		getParams = append(getParams, fmt.Sprintf("fromDate=%s", r.FromDate.Format(time.RFC3339)))
	}
	if !r.ToDate.IsZero() {
		getParams = append(getParams, fmt.Sprintf("toDate=%s", r.ToDate.Format(time.RFC3339)))
	}

	requestURL := fmt.Sprintf("%s/%s?%s",
		NukiApi,
		fmt.Sprintf(NukiLogEndpoint, r.SmartlockID),
		strings.Join(getParams, "&"))

	log.Debug().
		Str("request_url", requestURL).
		Send()

	httpReq, err := http.NewRequest(http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, err
	}
	httpReq.Header = http.Header{
		"Content-Type":  {"application/json"},
		"Authorization": {fmt.Sprintf("Bearer %s", r.Token)},
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

	var logResponses []model.NukiSmartlockLogResponse
	err = json.Unmarshal(body, &logResponses)
	if err != nil {
		return nil, err
	}

	return logResponses, err
}
