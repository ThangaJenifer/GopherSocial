package main

import (
	"net/http"
	"testing"
)

// ex 62 ***Test for not allowing unauthenticated users 401 ***
func TestGetUser(t *testing.T) {

	withRedis := config{
		redisCfg: redisConfig{
			enabled: true,
		},
	}
	//we need to create instance of application struct as we create new func for it and map all routes to it
	app := newTestApplication(t, withRedis)
	mux := app.mount()

	testToken, err := app.authenticator.GenerateToken(nil)
	if err != nil {
		t.Fatal(err)
	}

	//Test for not allowing unauthenticated users
	//using t.Run() this is like a subtest of main test
	t.Run("should not allow unauthenticated users", func(t *testing.T) {
		//check for 401 code
		req, err := http.NewRequest(http.MethodGet, "/v1/users/1", nil)
		if err != nil {
			//if err then fail the test using error
			t.Fatal(err)
		}
		//These both are pattern always we use alot so move this go to test_utlis.go and create a reusable excuteRequest function for it
		//	rr := httptest.NewRecorder()
		//	mux.ServeHTTP(rr, req)

		rr := excuteRequest(req, mux)
		//creating function checkResponseCode(t, http.StatusOK, rr.Code) in test_utils.go for repetative check
		// if rr.Code != http.StatusUnauthorized {
		// 	t.Errorf("expected the response code to be %d and we got %d", http.StatusUnauthorized, rr.Code)
		// }
		checkResponseCode(t, http.StatusUnauthorized, rr.Code)

	})

	t.Run("should allow authenticated requests", func(t *testing.T) {
		//check for 401 code
		req, err := http.NewRequest(http.MethodGet, "/v1/users/1", nil)
		if err != nil {
			//if err then fail the test using error
			t.Fatal(err)
		}

		req.Header.Set("Authorization", "Bearer "+testToken)
		//These both are pattern always we use alot so move this go to test_utlis.go and create a reusable excuteRequest function for it
		//	rr := httptest.NewRecorder()
		//	mux.ServeHTTP(rr, req)

		rr := excuteRequest(req, mux)

		checkResponseCode(t, http.StatusOK, rr.Code)
	})

}
