package main

import (
	"encoding/json"
	"expvar"
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

func (app *application) customVarHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	w.Write([]byte("{"))
	appendComma := false
	expvar.Do(func(kv expvar.KeyValue) {
		if kv.Key == "memstats" || kv.Key == "cmdline" {
			return // Skip memstats and cmdline
		}
		if appendComma {
			w.Write([]byte(","))
		} else {
			appendComma = true
		}

		b, _ := json.Marshal(kv.Key)
		w.Write(b)
		w.Write([]byte(":"))
		w.Write([]byte(kv.Value.String()))

	})
	w.Write([]byte("}"))
}
