package main

import (
	"github.com/justinas/alice"
	"net/http"
)

func (app *application) routes() http.Handler {
	base := alice.New(app.metrics, app.recoverPanic, app.enableCORS, app.rateLimit, app.authenticate)
	protected := alice.New(app.requireAuthenticatedUser, app.requireActivatedUser)
	mux := http.NewServeMux()

	mux.HandleFunc("OPTIONS /", app.preflightCORSHandler)

	mux.HandleFunc("GET /v1/healthcheck", app.healthcheckHandler)
	mux.HandleFunc("GET /debug/vars", app.customVarHandler)

	mux.Handle("GET /v1/movies", protected.Then(app.requirePermission("movies:read", app.listMoviesHandler)))
	mux.Handle("GET /v1/movies/{id}", protected.Then(app.requirePermission("movies:read", app.showMovieHandler)))
	mux.Handle("POST /v1/movies", protected.Then(app.requirePermission("movies:write", app.createMovieHandler)))
	mux.Handle("PATCH /v1/movies/{id}", protected.Then(app.requirePermission("movies:write", app.UpdateMovieHandler)))
	mux.Handle("DELETE /v1/movies/{id}", protected.Then(app.requirePermission("movies:write", app.DeleteMovieHandler)))

	mux.HandleFunc("POST /v1/users", app.registerUserHandler)
	mux.HandleFunc("PUT /v1/users/activated", app.activateUserHandler)

	mux.HandleFunc("POST /v1/tokens/activation", app.createActivationTokenHandler)
	mux.HandleFunc("POST /v1/tokens/authentication", app.createAuthenticationTokenHandler)

	return base.Then(mux)
}
