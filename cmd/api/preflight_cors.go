package main

import (
	"net/http"
)

func (app *application) preflightCORSHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Vary", "Access-Control-Request-Method")
	origin := r.Header.Get("Origin")
	if origin != "" && len(app.config.cors.trustedOrigins) != 0 {
		for i := range app.config.cors.trustedOrigins {
			if origin == app.config.cors.trustedOrigins[i] {
				if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
					w.Header().Set("Access-Control-Allow-Methods", "OPTIONS, PUT, PATCH, DELETE")
					w.Header().Set("Access-Control-Allow-Headers", "authorization, content-type")
					w.WriteHeader(http.StatusOK)
				}
			}
		}
	}
}
