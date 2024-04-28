package nukiapi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"nuki-logger/model"
)

const (
	NukiApi         = "https://api.nuki.io"
	NukiLogEndpoint = "smartlock/%s/log"
)

func ReadLogs(smartlockID, apiToken string) ([]model.NukiSmartlockLogResponse, error) {
	requestURL := fmt.Sprintf("%s/%s", NukiApi, fmt.Sprintf(NukiLogEndpoint, smartlockID))
	httpReq, err := http.NewRequest(http.MethodGet, requestURL+"?limit=30", nil)
	if err != nil {
		return nil, err
	}
	httpReq.Header = http.Header{
		"Content-Type":  {"application/json"},
		"Authorization": {fmt.Sprintf("Bearer %s", apiToken)},
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
