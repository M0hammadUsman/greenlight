package main

import (
	"fmt"
	"net/http"
)

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc(fmt.Sprint(http.MethodGet, " /v1/healthcheck"), app.healthcheckHandler)
	mux.HandleFunc(fmt.Sprint(http.MethodPost, " /v1/movies"), app.createMovieHandler)
	mux.HandleFunc(fmt.Sprint(http.MethodGet, " /v1/movies/{id}"), app.showMovieHandler)
	mux.HandleFunc(fmt.Sprint(http.MethodPatch, " /v1/movies/{id}"), app.UpdateMovieHandler)
	mux.HandleFunc(fmt.Sprint(http.MethodDelete, " /v1/movies/{id}"), app.DeleteMovieHandler)
	return mux
}
