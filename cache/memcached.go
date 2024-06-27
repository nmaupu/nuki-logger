package cache

import (
	"encoding/json"
	"errors"
	"time"

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

	if err := m.checkAlive(time.After(2 * time.Second)); err != nil {
		return err
	}

	return json.Unmarshal(item.Value, obj)
}

func (m memcached) Save(key string, obj any) error {
	bytes, err := json.Marshal(obj)
	if err != nil {
		return err
	}

	if err := m.checkAlive(time.After(2 * time.Second)); err != nil {
		return err
	}

	return m.Client.Set(&memcache.Item{
		Key:   key,
		Value: bytes,
	})
}

func (m memcached) checkAlive(timeout <-chan time.Time) error {
	// Check if server is alive
	tick := time.NewTicker(100 * time.Millisecond)
	defer tick.Stop()
	for {
		select {
		case <-timeout:
			return errors.New("unable to connect to the memcached server")
		case <-tick.C:
			if err := m.Client.Ping(); err == nil {
				return nil
			}
		}
	}
}
