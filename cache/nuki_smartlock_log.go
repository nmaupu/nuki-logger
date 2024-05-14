package cache

import (
	"github.com/nmaupu/nuki-logger/model"
)

const (
	nukiLoggerCacheFile = "/tmp/nuki-logger.cache"
)

func LoadCacheNukiSmartlockLogsFromDisk() ([]model.NukiSmartlockLogResponse, error) {
	var logResponses []model.NukiSmartlockLogResponse
	err := LoadCacheFromDisk(nukiLoggerCacheFile, &logResponses)
	if err != nil {
		return nil, err
	}
	return logResponses, err
}

func SaveCacheNukiSmartlockLogsToDisk(responses []model.NukiSmartlockLogResponse) error {
	return SaveCacheToDisk(responses, nukiLoggerCacheFile)
}
