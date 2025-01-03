package main

import (
	"net/http"
	"social/internal/store"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type userKey string

const userCtx userKey = "user"

// GetUser godoc
//
//	@Summary		Fetches a user profile
//	@Description	Fetches a user profile by ID
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"User ID"
//	@Success		200	{object}	store.User
//	@Failure		400	{object}	error
//	@Failure		404	{object}	error
//	@Failure		500	{object}	error
//	@Security		ApiKeyAuth
//	@Router			/users/{id} [get]
func (app *application) getUserHandler(w http.ResponseWriter, r *http.Request) {
	//user := getUserFromContext(r)
	//ex 59 caching needs ID from request
	userID, err := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
	if err != nil || userID < 1 {
		app.badRequestError(w, r, err)
		return
	}

	ctx := r.Context()
	//ex 59 using getUser func instead of app.store.Users.GetByID for caching with redis docker
	//commands for running docker and checking the keys are provided in store/cache/readme file to check
	//user, err := app.store.Users.GetByID(ctx, userID)
	user, err := app.getUser(ctx, userID)
	if err != nil {
		switch err {
		case store.ErrNotFound:
			app.notFoundError(w, r, err)
			return
		default:
			app.internalServerError(w, r, err)
			return
		}
	}
	err = app.jsonResponse(w, http.StatusOK, user)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

// FollowUser payload coming in for both follow and unfollow requests
// type FollowUser struct {
// 	UserID int64 `json:"user_id"`
// }

// FollowUser godoc
//
//	@Summary		Follows a user
//	@Description	Follows a user by ID
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			userID	path		int		true	"User ID"
//	@Success		204		{string}	string	"User followed"
//	@Failure		400		{object}	error	"User payload missing"
//	@Failure		404		{object}	error	"User not found"
//	@Security		ApiKeyAuth
//	@Router			/users/{userID}/follow [put]
//
// exercise 35
func (app *application) followUserHandler(w http.ResponseWriter, r *http.Request) {
	followerUser := getUserFromContext(r)
	followedID, err := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
	if err != nil {
		app.badRequestError(w, r, err)
		return
	}

	//Revert back to auth UserID from ctx
	// var payload FollowUser
	// if err := readJSON(w, r, &payload); err != nil {
	// 	app.badRequestError(w, r, err)
	// 	return
	// }
	ctx := r.Context()
	//creating a DB method to add followUserID to UserID
	err = app.store.Followers.Follow(ctx, followerUser.ID, followedID)
	if err != nil {
		switch err {
		case store.ErrConflict:
			//created a new error ErrConflict to get  a better error message who is consuming this api
			app.conflictResponse(w, r, err)
			return
		default:
			app.internalServerError(w, r, err)
			return
		}
	}

	err = app.jsonResponse(w, http.StatusNoContent, nil)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

// UnfollowUser gdoc
//
//	@Summary		Unfollow a user
//	@Description	Unfollow a user by ID
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			userID	path		int		true	"User ID"
//	@Success		204		{string}	string	"User unfollowed"
//	@Failure		400		{object}	error	"User payload missing"
//	@Failure		404		{object}	error	"User not found"
//	@Security		ApiKeyAuth
//	@Router			/users/{userID}/unfollow [put]
//
// exercise 35
func (app *application) unfollowUserHandler(w http.ResponseWriter, r *http.Request) {
	followerUser := getUserFromContext(r)
	unfollowedID, err := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
	if err != nil {
		app.badRequestError(w, r, err)
		return
	}
	ctx := r.Context()
	//creating a DB method to unfollow UserID for follower ID
	err = app.store.Followers.Unfollow(ctx, followerUser.ID, unfollowedID)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	err = app.jsonResponse(w, http.StatusNoContent, nil)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

// ActivateUser godoc
//
//	@Summary		Activates/Register a user
//	@Description	Activates/Register a user by invitation token
//	@Tags			users
//	@Produce		json
//	@Param			token	path		string	true	"Invitation token"
//	@Success		204		{string}	string	"User activated"
//	@Failure		404		{object}	error
//	@Failure		500		{object}	error
//	@Security		ApiKeyAuth
//	@Router			/users/activate/{token} [put]
//
// ex  45 user activation
func (app *application) activateUserHandler(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")

	err := app.store.Users.Activate(r.Context(), token)
	if err != nil {
		switch err {
		case store.ErrNotFound:
			app.notFoundError(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	err = app.jsonResponse(w, http.StatusNoContent, "")
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

}

// exercise 35
// func (app *application) userContextMiddleware(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		userID, err := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
// 		if err != nil {
// 			app.badRequestError(w, r, err)
// 			return
// 		}

// 		ctx := r.Context()

// 		user, err := app.store.Users.GetByID(ctx, userID)
// 		if err != nil {
// 			switch err {
// 			case store.ErrNotFound:
// 				app.badRequestError(w, r, err)
// 				return
// 			default:
// 				app.internalServerError(w, r, err)
// 				return
// 			}
// 		}

// 		ctx = context.WithValue(ctx, userCtx, user)
// 		//withContext returns a shallow copy of r with its context changed to ctx. The provided ctx must be non-nil
// 		next.ServeHTTP(w, r.WithContext(ctx))
// 	})
// }

func getUserFromContext(r *http.Request) *store.User {
	user, _ := r.Context().Value(userCtx).(*store.User)
	return user
}
