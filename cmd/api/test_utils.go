package main

import (
	"net/http"
	"net/http/httptest"
	"social/internal/auth"
	"social/internal/ratelimiter"
	"social/internal/store"
	"social/internal/store/cache"
	"testing"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// ex 62 we use this file to create all reusable functions for all tests
func newTestApplication(t *testing.T, cfg config) *application {
	t.Helper()

	//logger if we dont want no writting of output during testing use NewNop() method
	logger := zap.Must(zap.NewProduction()).Sugar()
	//logger := zap.NewNop().Sugar()
	mockStore := store.NewMockStore()
	mockCacheStore := cache.NewMockStore()
	testAuth := &auth.TestAuthenticator{}

	// Rate limiter
	rateLimiter := ratelimiter.NewFixedWindowRateLimiter(
		cfg.rateLimiter.RequestsPerTimeFrame,
		cfg.rateLimiter.TimeFrame,
	)

	return &application{
		logger:        logger,
		store:         mockStore,
		cacheStorage:  mockCacheStore,
		authenticator: testAuth,
		rateLimiter:   rateLimiter,
	}
}

func excuteRequest(req *http.Request, mux *chi.Mux) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	return rr
}

func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d", expected, actual)
	}
}
