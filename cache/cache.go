package cache

import "errors"

var (
	ErrCacheNoClient = errors.New("no client available")
)

type Cache interface {
	Load(key string, obj any) error
	Save(key string, data any) error
}
