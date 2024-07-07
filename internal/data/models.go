package data

import (
	"errors"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrRecordNotFound = errors.New("record not found")
)

type Models struct {
	Movies MovieModel
}

func NewModels(db *pgxpool.Pool) Models {
	return Models{MovieModel{db}}
}
