package main

import (
	"net/http"
	"social/internal/store"
)

// getUserFeedHandler godoc
//
//	@Summary		Fetches the user feed
//	@Description	Fetches the user feed
//	@Tags			feed
//	@Accept			json
//	@Produce		json
//	@Param			since	query		string	false	"Since"
//	@Param			until	query		string	false	"Until"
//	@Param			limit	query		int		false	"Limit"
//	@Param			offset	query		int		false	"Offset"
//	@Param			sort	query		string	false	"Sort"
//	@Param			tags	query		string	false	"Tags"
//	@Param			search	query		string	false	"Search"
//	@Success		200		{object}	[]store.PostWithMetadata
//	@Failure		400		{object}	error
//	@Failure		500		{object}	error
//	@Security		ApiKeyAuth
//	@Router			/users/feed [get]
//
// Exercise 37 User Feed Algorithm
func (app *application) getUserFeedHandler(w http.ResponseWriter, r *http.Request) {
	//Ex 38 pagination, filters, sort endpoint can look like /feed?limit=20&offset=0
	//we are using sliding window technique where offset for second set will be 10 and third will be 20 so on
	//Initialise with a default value of PaginatedFeedquery struct, then we will override with request parameters
	fq := store.PaginatedFeedQuery{
		Limit:  20,
		Offset: 0,
		Sort:   "desc",
	}
	//try to parse fq values present in request parameters
	fq, err := fq.Parse(r)
	if err != nil {
		app.badRequestError(w, r, err)
		return
	}

	err = Validate.Struct(fq)
	if err != nil {
		app.badRequestError(w, r, err)
		return
	}

	ctx := r.Context()
	//pass the feed query fq in GetUserFeed method
	feed, err := app.store.Posts.GetUserFeed(ctx, int64(12), fq)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusOK, feed); err != nil {
		app.internalServerError(w, r, err)
		return
	}

}
