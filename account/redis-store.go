package account

import (
	"encoding/json"
	"fmt"

	"github.com/go-redis/redis"
)

// RedisStore is an account store that is backed by Redis
type RedisStore struct {
	c redis.UniversalClient
}

// NewRedisStore creates an account store that is backed by Redis
func NewRedisStore(addresses []string) (*RedisStore, error) {
	rs := new(RedisStore)

	rs.c = redis.NewUniversalClient(&redis.UniversalOptions{
		Addrs: addresses,
	})

	_, err := rs.c.Ping().Result()
	if err != nil {
		return nil, err
	}

	return rs, nil
}

func (rs *RedisStore) Read(id string) (*Account, error) {
	a := Account{}

	val, err := rs.c.Get(id).Result()
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal([]byte(val), &a)
	if err != nil {
		return nil, err
	}

	if id != a.Username {
		return nil, fmt.Errorf("Username of account %s does not match %s", a.Username, id)
	}

	return &a, nil
}

func (rs *RedisStore) Write(a *Account) error {
	b, err := json.MarshalIndent(a, "", "  ")
	if err != nil {
		return err
	}

	return rs.c.Set(a.Username, string(b), 0).Err()
}
