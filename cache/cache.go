package cache

import (
	"encoding/json"
	"errors"
	"os"
)

func LoadCacheFromDisk(filename string, destObj any) error {
	if _, err := os.Stat(filename); errors.Is(err, os.ErrNotExist) {
		return nil
	}
	cache, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	return json.Unmarshal(cache, destObj)
}

func SaveCacheToDisk(data any, filename string) error {
	if data == nil {
		return nil
	}
	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return os.WriteFile(filename, bytes, 0644)
}
