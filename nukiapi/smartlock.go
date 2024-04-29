package nukiapi

import (
	"encoding/json"
	"fmt"
	"github.com/nmaupu/nuki-logger/model"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
)

type SmartlockReader struct {
	APICaller
	SmartlockID int64
}

func (r SmartlockReader) Execute() (*model.SmartlockResponse, error) {
	if r.SmartlockID == 0 {
		return nil, fmt.Errorf("smartlockid is mandatory")
	}
	if r.Token == "" {
		return nil, fmt.Errorf("token is mandatory")
	}

	requestURL := fmt.Sprintf("%s/%s",
		Api,
		fmt.Sprintf(SmartlockEndpoint, r.SmartlockID),
	)
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

	var response model.SmartlockResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	return &response, err
}
