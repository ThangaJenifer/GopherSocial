package ratelimiter

import (
	"sync"
	"time"
)

/*
if you want to be performance and you want to be distributed, consider  using Redis has your rate limiting.  you can have
very fast lookups that there instead of doing this in memory here, because  this is going to work for one server.
Because if you think about it here on the the drawing, we're just thinking about the one server setup.

checkout ex65RatelimitingWithRedis.png file in internal/store/ex-images

*/

type FixedWindowRateLimiter struct {
	sync.RWMutex                //using mutex as we have map of clients and need to make it concurrent access safe
	clients      map[string]int //each key is IP address and value is count of it tried the api
	limit        int
	window       time.Duration
}

func NewFixedWindowRateLimiter(limit int, window time.Duration) *FixedWindowRateLimiter {
	return &FixedWindowRateLimiter{
		clients: make(map[string]int),
		limit:   limit,
		window:  window,
	}
}

func (rl *FixedWindowRateLimiter) Allow(ip string) (bool, time.Duration) {
	rl.RLock()
	count, exists := rl.clients[ip]
	rl.RUnlock()

	if !exists || count < rl.limit {
		rl.Lock()
		if !exists {
			//if not exist, method resets the request count for a specific ip after the current time window ends.
			go rl.resetCount(ip)
		}
		rl.clients[ip]++
		rl.Unlock()
		//you can make other request, as he is allowed
		return true, 0
	}
	//he is not allowed and we need to return the window
	return false, rl.window
}

func (rl *FixedWindowRateLimiter) resetCount(ip string) {
	time.Sleep(rl.window)
	rl.Lock()
	delete(rl.clients, ip)
	rl.Unlock()
}
