package main

import (
	"context"
	"errors"
	"expvar"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"social/internal/env"
	"social/internal/mailer"
	"social/internal/store"
	"social/internal/store/cache"
	"syscall"
	"time"

	"social/internal/auth"

	"social/internal/ratelimiter"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.uber.org/zap"

	// "github.com/swaggo/swag/example/override/docs"
	"social/docs" //This is required to generate swagger docs
)

type application struct {
	config        config
	store         store.Storage
	cacheStorage  cache.Storage
	logger        *zap.SugaredLogger
	mailer        mailer.Client
	authenticator auth.Authenticator
	rateLimiter   ratelimiter.Limiter
}

type config struct {
	addr        string
	db          dbConfig
	env         string
	apiURL      string
	mail        mailConfig
	frontendURL string
	auth        authConfig
	redisCfg    redisConfig
	rateLimiter ratelimiter.Config
}

type redisConfig struct {
	addr    string
	pw      string
	db      int
	enabled bool
}

type authConfig struct {
	basic basicConfig
	token tokenConfig
}

type tokenConfig struct {
	secret string
	exp    time.Duration
	iss    string
}

type basicConfig struct {
	user string
	pass string
}

// ex 44 adding mailConfig expiry
type mailConfig struct {
	exp       time.Duration
	fromEmail string
	sendGrid  sendGridConfig
}

type sendGridConfig struct {
	apiKey string
}

type dbConfig struct {
	addr         string
	maxOpenConns int
	maxIdleConns int
	maxIdleTime  string
}

func (app *application) mount() *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	//ex65 to parse remoteaddr in RateLimiterMiddleware, we need to use this r.Use(middleware.RealIP) to parse ip address of the client
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Logger)
	//ex 66 CORS, placement of this is important above the ratelimiter middleware because it will be used by the server routes including rate limiter
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{env.GetString("CORS_ALLOWED_ORIGIN", "http://localhost:8080")}, // Use this to allow specific origin hosts
		//AllowedOrigins: []string{"https://*", "http://*"},
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))
	//ex65 rate limiter, we can put ratelimiter in all the requests like this at top of route for simplicity
	//else we can put ratelimiter for posts routes r.Use(app.RateLimiterMiddleware) inside that column
	r.Use(app.RateLimiterMiddleware)

	r.Use(middleware.Timeout(60 * time.Second))

	r.Route("/v1", func(r chi.Router) {
		//ex 50 basic auth, cleaner way to add middleware in chi r.With
		r.With(app.BasicAuthMiddleware()).Get("/health", app.healthCheckHandler)
		//ex 67 Server metrics
		r.With(app.BasicAuthMiddleware()).Get("/debug/vars", expvar.Handler().ServeHTTP)

		//Ex 40 Creating a swagger route under V1 to get documentation for Our APIs
		///swagger/* this route will have our documentation set
		docsURL := fmt.Sprintf("%s/swagger/doc.json", app.config.addr)
		r.Get("/swagger/*", httpSwagger.Handler(httpSwagger.URL(docsURL)))

		// /v1/posts/
		r.Route("/posts", func(r chi.Router) {
			//ex 52 using this as middleware for all below post routes
			r.Use(app.AuthTokenMiddleware)
			//POST /v1/posts
			r.Post("/", app.createPostHandler)
			//route for GET /v1/posts/{{postID}} reason we used postID
			//as we have more methods to filter out by postID like PATCH, Delete posts
			r.Route("/{postID}", func(r chi.Router) {
				//putting postsContextMiddleware here so it affects only to above route of posts ID
				r.Use(app.postsContextMiddleware)

				r.Get("/", app.getPostHandler)
				//exercise 28 updating and deleting handler
				//r.Delete("/", app.deletePostHandler)
				//r.Patch("/", app.updatePostHandler)
				//ex 56 Role base authorization using the middleware checkPostOwnership for update and delete handlers
				r.Patch("/", app.checkPostOwnership("moderator", app.updatePostHandler))
				r.Delete("/", app.checkPostOwnership("admin", app.deletePostHandler))

			})
		})

		// /v1/users
		r.Route("/users", func(r chi.Router) {
			//ex 45 User Activation
			r.Put("/activate/{token}", app.activateUserHandler)

			//Get for profile fetching exercise 34
			r.Route("/{userID}", func(r chi.Router) {
				//ex 52 using this as middleware for all below accessing user by ID routes
				r.Use(app.AuthTokenMiddleware)
				r.Get("/", app.getUserHandler)
				// need route PUT /v1/users/42/follow for follow. exercise 35
				//we can use route DELETE /v1/users/42/follow for unfollow. But we use same PUT for follow and unfollow
				r.Put("/follow", app.followUserHandler)
				r.Put("/unfollow", app.unfollowUserHandler)
			})
			//creating user feed like we have on facebook/instagram exercise 37 v1/users/12/feed who is userID we want
			r.Group(func(r chi.Router) {
				r.Use(app.AuthTokenMiddleware)
				r.Get("/feed", app.getUserFeedHandler)
			})
		})

		//public routes
		//exe 43 this is used for user authentication so a public route
		r.Route("/authentication", func(r chi.Router) {
			r.Post("/user", app.registerUserHandler)
			//ex 51 JWT generate token
			r.Post("/token", app.createTokenHandler)
		})

	})

	return r
}

func (app *application) run(mux *chi.Mux) error {
	//Docs
	docs.SwaggerInfo.Version = version
	docs.SwaggerInfo.Host = app.config.apiURL
	docs.SwaggerInfo.BasePath = "/v1"

	srv := &http.Server{
		Addr:         app.config.addr,
		Handler:      mux,
		WriteTimeout: time.Second * 30,
		ReadTimeout:  time.Second * 10,
		IdleTimeout:  time.Minute,
	}
	//ex 17 graceful server shutdown
	shutdown := make(chan error)

	go func() {
		// For a channel used for notification of just one signal value, a buffer of size 1 is sufficient.
		quit := make(chan os.Signal, 1)
		//Notify causes package signal to relay incoming signals to channel quit.
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		s := <-quit

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		app.logger.Infow("signal caught", "signal", s.String())

		shutdown <- srv.Shutdown(ctx)
	}()

	app.logger.Infow("Server has started", "addr", app.config.addr, "env", app.config.env)

	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	err = <-shutdown
	if err != nil {
		return err
	}

	app.logger.Infow("server has stopped", "addr", app.config.addr, "env", app.config.env)

	return nil
}
