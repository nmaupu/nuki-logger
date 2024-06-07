package cache

import (
	"encoding/json"

	"github.com/bradfitz/gomemcache/memcache"
)

var _ Cache = (*memcached)(nil)

type memcached struct {
	*memcache.Client
}

func NewMemcached(servers []string) Cache {
	return &memcached{
		Client: memcache.New(servers...),
	}
}

func (m memcached) Load(key string, obj any) error {
	item, err := m.Client.Get(key)
	if err != nil {
		return err
	}

	return json.Unmarshal(item.Value, obj)
}

func (m memcached) Save(key string, obj any) error {
	bytes, err := json.Marshal(obj)
	if err != nil {
		return err
	}

	return m.Client.Set(&memcache.Item{
		Key:   key,
		Value: bytes,
	})
}
