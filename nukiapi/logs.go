package nukiapi

import (
	"encoding/json"
	"fmt"
	"github.com/nmaupu/nuki-logger/model"
	"strings"
	"time"
)

type LogsReader struct {
	APICaller
	SmartlockID int64
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
		Api,
		fmt.Sprintf(LogsEndpoint, r.SmartlockID),
		strings.Join(getParams, "&"))

	body, err := r.execAPIGet(requestURL)
	if err != nil {
		return nil, err
	}

	var responses []model.NukiSmartlockLogResponse
	err = json.Unmarshal(body, &responses)
	return responses, err
}
