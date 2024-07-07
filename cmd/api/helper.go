package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

func (app *application) readIDParam(r *http.Request) (int64, error) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, errors.New("invalid id param")
	}
	return id, nil
}

type envelop map[string]any

func (app *application) writeJSON(w http.ResponseWriter, data envelop, status int, headers http.Header) error {
	hello, err := json.MarshalIndent(data, "", "\t")
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

// readJSON reads the next JSON-encoded value from the request and stores it in the value pointed to by dst
func (app *application) readJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	maxBytes := 1_048_576
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		var syntaxErr *json.SyntaxError
		var unmarshalTypeErr *json.UnmarshalTypeError
		var invalidUnmarshalErr *json.InvalidUnmarshalError

		switch {
		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly-formed JSON")
		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")
		case errors.As(err, &syntaxErr):
			return fmt.Errorf("body contains badly-formed JSON (at character %d)", syntaxErr.Offset)
		case errors.As(err, &unmarshalTypeErr):
			if unmarshalTypeErr.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeErr.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type (at character %d)", unmarshalTypeErr.Offset)
		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.Trim(err.Error(), "json: unknown field ")
			return fmt.Errorf("body contains unknown key %s", fieldName)
		case err.Error() == "http: request body too large":
			return fmt.Errorf("body must not be larger than %d bytes", maxBytes)
		case errors.As(err, &invalidUnmarshalErr):
			panic(err)
		default:
			return err
		}
	}

	if err := dec.Decode(&struct{}{}); err != io.EOF {
		return errors.New("body must only contain a single JSON value")
	}
	return nil
}
