package main

import (
	"github.com/justinas/alice"
	"net/http"
)

func (app *application) routes() http.Handler {
	middlewares := alice.New(app.recoverPanic, app.rateLimit)
	mux := http.NewServeMux()

	mux.HandleFunc("GET /v1/healthcheck", app.healthcheckHandler)

	mux.HandleFunc("POST /v1/movies", app.createMovieHandler)
	mux.HandleFunc("GET /v1/movies/{id}", app.showMovieHandler)
	mux.HandleFunc("PATCH /v1/movies/{id}", app.UpdateMovieHandler)
	mux.HandleFunc("DELETE /v1/movies/{id}", app.DeleteMovieHandler)
	mux.HandleFunc("GET /v1/movies", app.listMoviesHandler)

	mux.HandleFunc("POST /v1/users", app.registerUserHandler)
	
	return middlewares.Then(mux)
}
