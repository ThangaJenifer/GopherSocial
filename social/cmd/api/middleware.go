package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"social/internal/store"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// ex 52, middleware to plug into routers for validating tokens
func (app *application) AuthTokenMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		//read the auth header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			app.unauthorizedErrorResponse(w, r, fmt.Errorf("authorization header is missing"))
			return
		}

		//parse it -> get the base64
		parts := strings.Split(authHeader, " ")
		//Authorization: Bearer HEWRKkkjasdjfjh; last part is the token, so lets write validation
		if len(parts) != 2 || parts[0] != "Bearer" {
			app.unauthorizedErrorResponse(w, r, fmt.Errorf("authorization header is malformed"))
			return
		}
		//This token is a jwt token string
		token := parts[1]
		jwtToken, err := app.authenticator.ValidateToken(token)
		if err != nil {
			app.unauthorizedErrorResponse(w, r, err)
			return
		}

		claims, _ := jwtToken.Claims.(jwt.MapClaims)

		userID, err := strconv.ParseInt(fmt.Sprintf("%.f", claims["sub"]), 10, 64)
		if err != nil {
			app.unauthorizedErrorResponse(w, r, err)
			return
		}

		ctx := r.Context()
		//ex 59, we fetch the user profile for every authenticated user request , this is right place to cache the performance of the user
		//instead of doing it from the getUserHandler method. so lets implement cache on this layer.
		//Lets create a function for cache which will abstract this way for a consumer
		// user, err := app.store.Users.GetByID(ctx, userID)
		// if err != nil {
		// 	app.unauthorizedErrorResponse(w, r, err)
		// 	return
		// }
		user, err := app.getUser(ctx, userID)
		if err != nil {
			app.unauthorizedErrorResponse(w, r, err)
			return
		}

		//now lets set the user variable into the context by creating a new context
		ctx = context.WithValue(ctx, userCtx, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ex 50
func (app *application) BasicAuthMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			//read the auth header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				app.unauthorizedBasicErrorResponse(w, r, fmt.Errorf("authorization header is missing"))
				return
			}
			//parse it -> get the base64
			parts := strings.Split(authHeader, " ")
			//Authorization: Basic HEWRKkkjasdjfjh; last part is the token, so lets write validation
			if len(parts) != 2 || parts[0] != "Basic" {
				app.unauthorizedBasicErrorResponse(w, r, fmt.Errorf("authorization header is malformed"))
				return
			}

			//decode it using base64 package
			decoded, err := base64.StdEncoding.DecodeString(parts[1])
			if err != nil {
				app.unauthorizedBasicErrorResponse(w, r, err)
				return

			}
			//check the credentials
			username := app.config.auth.basic.user
			pass := app.config.auth.basic.pass

			//we are just spliting it by : and taking maximum 2 substrings, username:passowrd
			//sometimes string can be username:passowrdextra:123 so it will divide only username and passwordextra:123
			creds := strings.SplitN(string(decoded), ":", 2)
			//validating if it has username and passowrd
			if len(creds) != 2 || creds[0] != username || creds[1] != pass {
				app.unauthorizedBasicErrorResponse(w, r, fmt.Errorf("invalid credentials"))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// ex 56 This is authorization for the posts
func (app *application) checkPostOwnership(requiredRole string, next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//getting authenticated user
		user := getUserFromContext(r)
		post := getPostFromCtx(r)

		//if it is the users post, if user is owner of post then he can go to update/delete handler
		if post.UserID == user.ID {
			next.ServeHTTP(w, r)
			return
		}

		//role precedence check
		allowed, err := app.checkRolePrecedence(r.Context(), user, requiredRole)
		if err != nil {
			app.internalServerError(w, r, err)
			return
		}
		//if allowed is false then it should not allow the user to do operations, so forbiddenResponse
		if !allowed {
			app.forbiddenResponse(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// ex 56
func (app *application) checkRolePrecedence(ctx context.Context, user *store.User, roleName string) (bool, error) {
	//fetch the role
	role, err := app.store.Roles.GetByName(ctx, roleName)
	if err != nil {
		return false, err
	}
	//We will add Role in userModels, this will check if passed roleName Moderator/admin is greater less/greater than user role
	return user.Role.Level >= role.Level, nil
}

// ex 59
func (app *application) getUser(ctx context.Context, userID int64) (*store.User, error) {
	//if cache is not enabled then fetch from postgres database
	if !app.config.redisCfg.enabled {
		return app.store.Users.GetByID(ctx, userID)
	}

	app.logger.Infow("cache hit", "key", "user", "ID", userID)
	//retrive from redis cache
	user, err := app.cacheStorage.Users.Get(ctx, userID)
	if err != nil {
		return nil, err
	}
	//if this user is empty then fetch from postgres database and update the redis cache
	if user == nil {
		app.logger.Infow("fetching from DB", "id", userID)
		user, err = app.store.Users.GetByID(ctx, userID)
		if err != nil {
			return nil, err
		}

		err = app.cacheStorage.Users.Set(ctx, user)
		if err != nil {
			return nil, err
		}
	}

	return user, nil
}

// ex 65 RateLimiterMiddleware
func (app *application) RateLimiterMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if app.config.rateLimiter.Enabled {
			if allow, retryAfter := app.rateLimiter.Allow(r.RemoteAddr); !allow {
				//if allow is false created new error under errors.go file rateLimitExceededResponse
				app.rateLimitExceededResponse(w, r, retryAfter.String())
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}
