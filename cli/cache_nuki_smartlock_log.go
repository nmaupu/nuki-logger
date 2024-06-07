package cli

import (
	"errors"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/nmaupu/nuki-logger/cache"
	"github.com/nmaupu/nuki-logger/model"
)

const (
	cacheNukiSmartlockLogsKey = "nuki-smartlock-logs"
)

type CacheNukiSmartlockLogs struct {
	Client cache.Cache
}

func (c CacheNukiSmartlockLogs) Load() ([]model.NukiSmartlockLogResponse, error) {
	if c.Client == nil {
		return nil, cache.ErrCacheNoClient
	}

	logResponses := []model.NukiSmartlockLogResponse{}
	if err := c.Client.Load(cacheNukiSmartlockLogsKey, &logResponses); err != nil {
		switch {
		case errors.Is(err, memcache.ErrCacheMiss):
			return []model.NukiSmartlockLogResponse{}, nil
		case errors.Is(err, memcache.ErrNoServers):
			return []model.NukiSmartlockLogResponse{}, nil
		default:
			return nil, err
		}
	}
	return logResponses, nil
}

func (c CacheNukiSmartlockLogs) Save(responses []model.NukiSmartlockLogResponse) error {
	if c.Client == nil {
		return cache.ErrCacheNoClient
	}
	return c.Client.Save(cacheNukiSmartlockLogsKey, responses)
}
