package ratelimiter

import "time"

//ex 65 rate limiter
type Limiter interface {
	//Allow method which will have a ip address where how much time to wait will be time.Duration
	Allow(ip string) (bool, time.Duration)
}

//config for rate limiter, as RequestPerTimeFrame can be 10 for TimeFrame of 1 hour such that
type Config struct {
	RequestsPerTimeFrame int
	TimeFrame            time.Duration
	Enabled              bool
}

//200 ok

//429 some requests in future, how much time we need to wait until we make other request
//set a header like Make another request in 40 seconds
