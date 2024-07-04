package main

import (
	"log/slog"
	"net/http"
)

func (app *application) healthcheckHandler(w http.ResponseWriter, _ *http.Request) {
	data := map[string]string{
		"status":      "available",
		"environment": app.config.env,
		"version":     version,
	}
	err := app.writeJSON(w, data, 200, nil)
	if err != nil {
		slog.Error(err.Error())
		http.Error(w, "The server encountered a problem and could not process your request", http.StatusInternalServerError)
	}
}
