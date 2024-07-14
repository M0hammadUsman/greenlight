package main

import (
	"errors"
	"github.com/M0hammadUsman/greenlight/internal/data"
	"github.com/M0hammadUsman/greenlight/internal/validator"
	"log/slog"
	"net/http"
	"time"
)

func (app *application) createActivationTokenHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email string `json:"email"`
	}
	if err := app.readJSON(w, r, &input); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	v := validator.New()
	if data.ValidateEmail(v, input.Email); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	usr, err := app.models.Users.GetByEmail(input.Email)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			v.AddError("email", "no matching email address found")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	if usr.Activated {
		v.AddError("email", "user has already been activated")
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	token, err := app.models.Tokens.New(usr.ID, 3*24*time.Hour, data.ScopeActivation)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	app.runInBackground(func() {
		d := map[string]any{"activationToken": token.PlainText}
		if err = app.mailer.Send(usr.Email, "token_activation.tmpl.html", d); err != nil {
			slog.Error(err.Error())
		}
	})
	env := envelop{"message": "an email will be sent to you containing activation instructions"}
	if err = app.writeJSON(w, env, http.StatusAccepted, nil); err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
