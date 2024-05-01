package nukiapi

import (
	"encoding/json"
	"fmt"
	"github.com/nmaupu/nuki-logger/model"
)

type SmartlockAuthReader struct {
	APICaller
	SmartlockID int64
}

func (r SmartlockAuthReader) Execute() ([]model.SmartlockAuthResponse, error) {
	if r.SmartlockID == 0 {
		return nil, fmt.Errorf("smartlockid is mandatory")
	}
	if r.Token == "" {
		return nil, fmt.Errorf("token is mandatory")
	}

	requestURL := fmt.Sprintf("%s/%s", Api, fmt.Sprintf(SmartlockAuthEndpoint, r.SmartlockID))
	body, err := r.execAPIGet(requestURL)
	if err != nil {
		return nil, err
	}

	var responses []model.SmartlockAuthResponse
	err = json.Unmarshal(body, &responses)
	return responses, err
}
