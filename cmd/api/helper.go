package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
)

func (app *application) readIDParam(r *http.Request) (int64, error) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, errors.New("invalid id param")
	}
	return id, nil
}

func (app *application) writeJSON(w http.ResponseWriter, data any, status int, headers http.Header) error {
	hello, err := json.Marshal(data)
	if err != nil {
		return err
	}
	hello = append(hello, '\n')
	// At this point, we know that we won't encounter any more errors before writing the
	// response, so it's safe to add any headers that we want to include
	for k, v := range headers {
		w.Header()[k] = v
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(hello)
	return nil
}
