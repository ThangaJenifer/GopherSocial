package main

import (
	"context"
	"errors"
	"net/http"
	"social/internal/store"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type postKey string

const postCtx postKey = "post"

/*
I wouldn't say very, but there is something bad here.And the first one is this. Now, I started with this because this is what most people will think about it. And I want to show you that this is not the most appropriate way to do it.
Because if you think about it, if we accept a post, this is basically telling the user that we are going to accept the post, right? So everything inside of this data structure, we are accepting and meaning that we are accepting credit. That's data that we're going to accept. User ID what does this mean is that the user can send this data and overwrite the data so he can corrupt and make unintended changes to our data storage.
And this is not what we want. So the best way to do this is to create a style, a type a structure that just has the post payload. So the create post payload, which I'm going to create here
*/
// Create a structure which has only the post payload
type CreatePostPayload struct {
	Title   string   `json:"title" validate:"required,max=100"`
	Content string   `json:"content" validate:"required,max=1000"`
	Tags    []string `json:"tags"`
}

// CreatePost godoc
//
//	@Summary		Creates a post
//	@Description	Creates a post
//	@Tags			posts
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		CreatePostPayload	true	"Post payload"
//	@Success		201		{object}	store.Post
//	@Failure		400		{object}	error
//	@Failure		401		{object}	error
//	@Failure		500		{object}	error
//	@Security		ApiKeyAuth
//	@Router			/posts [post]
func (app *application) createPostHandler(w http.ResponseWriter, r *http.Request) {
	var payload CreatePostPayload
	err := readJSON(w, r, &payload)
	if err != nil {
		app.badRequestError(w, r, err)
		return
	}

	/*
		we can use validation as below , or create a method for payload validation like ValidatePayload()
		but we use json tag method to validate it, which is simplier
		if payload.Content == "" {
			app.badRequestError(w, r, fmt.Errorf("content is required"))
			return
		}
	*/
	//we use Validate.Struct method to validate struct with conditons mentioned
	//learn usage https://github.com/go-playground/validator
	if err := Validate.Struct(payload); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	user := getUserFromContext(r)

	post := &store.Post{
		Title:   payload.Title,
		Content: payload.Content,
		Tags:    payload.Tags,
		UserID:  user.ID,
	}

	ctx := r.Context()

	err = app.store.Posts.Create(ctx, post)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	err = app.jsonResponse(w, http.StatusCreated, post)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

}

// GetPost godoc
//
//	@Summary		Fetches a post
//	@Description	Fetches a post by ID
//	@Tags			posts
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"Post ID"
//	@Success		200	{object}	store.Post
//	@Failure		404	{object}	error
//	@Failure		500	{object}	error
//	@Security		ApiKeyAuth
//	@Router			/posts/{id} [get]
func (app *application) getPostHandler(w http.ResponseWriter, r *http.Request) {
	post := getPostFromCtx(r)

	// idParam := chi.URLParam(r, "postID")
	// id, err := strconv.ParseInt(idParam, 10, 64)
	// if err != nil {
	// 	app.internalServerError(w, r, err)
	// 	return
	// }
	// ctx := r.Context()

	// post, err := app.store.Posts.GetByID(ctx, id)
	// if err != nil {
	// 	switch {
	// 	case errors.Is(err, store.ErrNotFound):
	// 		app.notFoundError(w, r, err)
	// 	default:
	// 		app.internalServerError(w, r, err)
	// 	}
	// 	return
	// }

	//Exercise 27 everytime we fetch post lets fetch its comments as well
	comments, err := app.store.Comments.GetByPostID(r.Context(), post.ID)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	post.Comments = comments

	err = app.jsonResponse(w, http.StatusOK, post)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

}

// DeletePost godoc
//
//	@Summary		Deletes a post
//	@Description	Delete a post by ID
//	@Tags			posts
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"Post ID"
//	@Success		204	{object} string
//	@Failure		404	{object}	error
//	@Failure		500	{object}	error
//	@Security		ApiKeyAuth
//	@Router			/posts/{id} [delete]
//
// excerise 28 deleting and updating post
func (app *application) deletePostHandler(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "postID")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}
	ctx := r.Context()

	if err := app.store.Posts.Delete(ctx, id); err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			app.notFoundError(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}
	//using no content here as not returning anything
	w.WriteHeader(http.StatusNoContent)
}

/*
we can send a payload like { "title": "new title"} where content is not intialised, In this case we
want only title to be changed and not content. We can do vice-versa as well.
So if content = "" if we take pointer to it then it is null. So while updating we will take
pointers so that can be nullable *string. default value of pointer empty string in go is null
*/
type UpdatePostPayload struct {
	Title   *string `json:"title" validate:"omitempty,max=1000`
	Content *string `json:"content" validate:"omitempty,max=1000`
}

// UpdatePost godoc
//
//	@Summary		Updates a post
//	@Description	Updates a post by ID
//	@Tags			posts
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int					true	"Post ID"
//	@Param			payload	body		UpdatePostPayload	true	"Post payload"
//	@Success		200		{object}	store.Post
//	@Failure		400		{object}	error
//	@Failure		401		{object}	error
//	@Failure		404		{object}	error
//	@Failure		500		{object}	error
//	@Security		ApiKeyAuth
//	@Router			/posts/{id} [patch]
func (app *application) updatePostHandler(w http.ResponseWriter, r *http.Request) {
	post := getPostFromCtx(r)

	var payload UpdatePostPayload
	err := readJSON(w, r, &payload)
	if err != nil {
		app.badRequestError(w, r, err)
	}

	err = Validate.Struct(payload)
	if err != nil {
		app.badRequestError(w, r, err)
		return
	}
	//Used that nullable concept and checked the value
	if payload.Content != nil {
		post.Content = *payload.Content
	}

	if payload.Title != nil {
		post.Title = *payload.Title
	}

	if err := app.store.Posts.Update(r.Context(), post); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusOK, post); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

// Excersie 28 this will fetch the post and foing to put into the context which will be middleware
// so it will write to every handler which we want to put it in
func (app *application) postsContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idParam := chi.URLParam(r, "postID")
		id, err := strconv.ParseInt(idParam, 10, 64)
		if err != nil {
			app.internalServerError(w, r, err)
			return
		}
		ctx := r.Context()

		post, err := app.store.Posts.GetByID(ctx, id)
		if err != nil {
			switch {
			case errors.Is(err, store.ErrNotFound):
				app.notFoundError(w, r, err)
			default:
				app.internalServerError(w, r, err)
			}
			return
		}

		//Excersie 28 we need to use previous context ctx := r.Context() and create new context and
		//insert our post in it. We never mutate context but always create a new one from scratch
		//should not use basic type untyped string as key in context.WithValue not a best pratice to avoid collusions
		//ctx = context.WithValue(ctx, "post", post)
		ctx = context.WithValue(ctx, postCtx, post)

		//we send request.WithContext and send in the newly created context having post data
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getPostFromCtx(r *http.Request) *store.Post {
	post, _ := r.Context().Value(postCtx).(*store.Post)
	return post
}
