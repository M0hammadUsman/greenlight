package main

import (
	"context"
	"github.com/M0hammadUsman/greenlight/internal/data"
	"net/http"
)

type contextKey string

// Just to be safe from key collision in Request Contexts
const userContextKey = contextKey("user")

func (app *application) contextSetUser(r *http.Request, user *data.User) *http.Request {
	ctx := context.WithValue(r.Context(), userContextKey, user)
	return r.WithContext(ctx)
}

func (app *application) contextGetUser(r *http.Request) *data.User {
	user, ok := r.Context().Value(userContextKey).(*data.User)
	if !ok {
		panic("missing user value in the request context")
	}
	return user
}
