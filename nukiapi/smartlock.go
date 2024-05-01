package nukiapi

import (
	"encoding/json"
	"fmt"
	"github.com/nmaupu/nuki-logger/model"
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

	requestURL := fmt.Sprintf("%s/%s", Api, fmt.Sprintf(SmartlockEndpoint, r.SmartlockID))
	body, err := r.execAPIGet(requestURL)
	if err != nil {
		return nil, err
	}

	var response model.SmartlockResponse
	err = json.Unmarshal(body, &response)
	return &response, err
}
