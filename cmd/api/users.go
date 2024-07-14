package main

import (
	"errors"
	"github.com/M0hammadUsman/greenlight/internal/data"
	"github.com/M0hammadUsman/greenlight/internal/validator"
	"log/slog"
	"net/http"
	"time"
)

func (app *application) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := app.readJSON(w, r, &input); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	user := &data.User{
		Name:      input.Name,
		Email:     input.Email,
		Activated: false,
	}
	if err := user.Password.Set(input.Password); err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	v := validator.New()
	if data.ValidateUser(v, user); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	if err := app.models.Users.Insert(user); err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateEmail):
			v.AddError("email", "a user with this email address already exists")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	} // Send welcome email
	token, err := app.models.Tokens.New(user.ID, 3*24*time.Hour, data.ScopeActivation)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
	app.runInBackground(func() {
		d := map[string]any{
			"activationToken": token.PlainText,
			"userID":          user.ID,
		}
		if err = app.mailer.Send(user.Email, "user_welcome.tmpl.html", d); err != nil {
			slog.Error(err.Error())
		}
	})
	if err = app.writeJSON(w, envelop{"user": user}, http.StatusAccepted, nil); err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

func (app *application) activateUserHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		TokenPlainText string `json:"token"`
	}
	if err := app.readJSON(w, r, &input); err != nil {
		app.badRequestResponse(w, r, err)
	}
	v := validator.New()
	if data.ValidateTokenPlainText(v, input.TokenPlainText); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	usr, err := app.models.Users.GetForToken(data.ScopeActivation, input.TokenPlainText)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			v.AddError("token", "invalid or expired activation token")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	usr.Activated = true
	if err = app.models.Users.UpdateUser(usr); err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	if err = app.models.Tokens.DeleteAllForUser(data.ScopeActivation, usr.ID); err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	if err = app.writeJSON(w, envelop{"user": usr}, http.StatusOK, nil); err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
