package main

import (
	"net/http"
)

func (app *application) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	env := envelop{
		"status": "available",
		"systemInfo": map[string]string{
			"environment": app.config.env,
			"version":     version,
		},
	}
	err := app.writeJSON(w, env, http.StatusOK, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
