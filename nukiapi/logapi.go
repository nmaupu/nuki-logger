package nukiapi

import (
	"encoding/json"
	"fmt"
	"github.com/nmaupu/nuki-logger/model"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	NukiApi         = "https://api.nuki.io"
	NukiLogEndpoint = "smartlock/%d/log"
)

type LogsReader struct {
	SmartlockID int64
	Token       string
	Limit       int
	FromDate    time.Time
	ToDate      time.Time
}

func (r LogsReader) Execute() ([]model.NukiSmartlockLogResponse, error) {
	if r.SmartlockID == 0 {
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

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("error while querying Nuki API (status: %s): %s", resp.Status, string(body))
	}

	var logResponses []model.NukiSmartlockLogResponse
	err = json.Unmarshal(body, &logResponses)
	if err != nil {
		return nil, err
	}

	return logResponses, err
}
