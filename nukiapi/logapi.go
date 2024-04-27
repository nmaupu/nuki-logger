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

func ReadLog(smartlockID, apiToken string) error {
	requestURL := fmt.Sprintf("%s/%s", NukiApi, fmt.Sprintf(NukiLogEndpoint, smartlockID))
	httpReq, err := http.NewRequest(http.MethodGet, requestURL+"?limit=50", nil)
	if err != nil {
		return err
	}
	httpReq.Header = http.Header{
		"Content-Type":  {"application/json"},
		"Authorization": {fmt.Sprintf("Bearer %s", apiToken)},
	}
	client := http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var logResponses []model.NukiSmartlockLogResponse
	err = json.Unmarshal(body, &logResponses)
	if err != nil {
		return err
	}

	for _, logResponse := range logResponses {
		if logResponse.Source == model.NukiSourceKeypadCode {
			fmt.Printf("%s - deviceType=%s, action=%s, trigger=%s, state=%s, source=%s, name=%s\n",
				logResponse.Date,
				logResponse.DeviceType,
				logResponse.Action,
				logResponse.Trigger,
				logResponse.State,
				logResponse.Source,
				logResponse.Name,
			)
		} else {
			fmt.Printf("%s - deviceType=%s, action=%s, trigger=%s, state=%s, source=%s\n",
				logResponse.Date,
				logResponse.DeviceType,
				logResponse.Action,
				logResponse.Trigger,
				logResponse.State,
				logResponse.Source,
			)
		}
	}

	return nil
}
