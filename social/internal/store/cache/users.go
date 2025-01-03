package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"social/internal/store"
	"time"

	"github.com/go-redis/redis/v8"
)

// ex 59
type UserStore struct {
	rdb *redis.Client
}

const UserExpTime = time.Minute

// ex 59
func (s *UserStore) Get(ctx context.Context, userID int64) (*store.User, error) {
	// {"user-42": "42"}, Redis is key-value store. We need good key to retrieve our users. we need to make key concrete like user-42
	//but not using just user which is more abstract, add user-id like this so that whole key will be unique otherwise we will have conflict with cache.
	cacheKey := fmt.Sprintf("user-%v", userID)

	data, err := s.rdb.Get(ctx, cacheKey).Result()
	if err == redis.Nil { // when fetching first time for redis, key wont be present and we get error redis.Nil, so we use this to avoid getting app crash, we use this
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	var user store.User
	if data != "" {
		err := json.Unmarshal([]byte(data), &user)
		if err != nil {
			return nil, err
		}
	}

	return &user, nil
}

// ex 59
func (s *UserStore) Set(ctx context.Context, user *store.User) error {
	cacheKey := fmt.Sprintf("user-%v", user.ID)

	json, err := json.Marshal(user)
	if err != nil {
		return err
	}
	//using method Setexpiration to use TTL timetolive concept
	err = s.rdb.SetEX(ctx, cacheKey, json, UserExpTime).Err()
	if err != nil {
		return err
	}

	return nil

}
