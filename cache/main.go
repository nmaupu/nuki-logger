package cache

import (
	"encoding/json"
	"errors"
	"nuki-logger/model"
	"os"
)

const (
	cacheFile = "/tmp/nuki-logger.cache"
)

func LoadCacheFromDisk() ([]model.NukiSmartlockLogResponse, error) {
	if _, err := os.Stat(cacheFile); errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}

	cache, err := os.ReadFile(cacheFile)
	if err != nil {
		return nil, err
	}

	var logResponses []model.NukiSmartlockLogResponse
	err = json.Unmarshal(cache, &logResponses)
	if err != nil {
		return nil, err
	}
	return logResponses, nil
}

func SaveCacheToDisk(responses []model.NukiSmartlockLogResponse) error {
	if responses == nil {
		return nil
	}
	bytes, err := json.Marshal(responses)
	if err != nil {
		return err
	}
	return os.WriteFile(cacheFile, bytes, 0644)
}
