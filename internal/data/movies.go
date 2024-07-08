package data

import (
	"errors"
	"github.com/M0hammadUsman/greenlight/internal/validator"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"strings"
	"time"
)

type Movie struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"-"`
	Title     string    `json:"title"`
	Year      int32     `json:"year,omitempty,string"`
	Runtime   Runtime   `json:"runtime,omitempty"`
	Genres    []string  `json:"genres,omitempty"`
	Version   int32     `json:"version"`
}

func ValidateMovie(v *validator.Validator, movie *Movie) {
	v.Check(strings.TrimSpace(movie.Title) != "", "title", "must be provided & not blank")
	v.Check(len(movie.Title) <= 500, "title", "must not be more than 500 bytes long")

	v.Check(movie.Year != 0, "year", "must be provided")
	v.Check(movie.Year >= 1888, "year", "must be greater than 1888")
	v.Check(movie.Year <= int32(time.Now().Year()), "year", "must not be in the future")

	v.Check(movie.Runtime != 0, "runtime", "must be provided !0")
	v.Check(movie.Runtime > 0, "runtime", "must be a positive integer")

	v.Check(movie.Genres != nil, "genres", "must be provided")
	v.Check(len(movie.Genres) >= 1, "genres", "must contain at least 1 genre")
	v.Check(len(movie.Genres) <= 5, "genres", "must not contain more than 5 genres")
	v.Check(validator.Unique(movie.Genres), "genres", "must not contain duplicate values")
}

type MovieModel struct {
	DB *pgxpool.Pool
}

func (m MovieModel) Insert(movie *Movie) error {
	query := `
		INSERT INTO movies (title, year, runtime, genres)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, version
		`
	args := []any{movie.Title, movie.Year, movie.Runtime, movie.Genres}
	ctx, cancel := newQueryContext(3)
	defer cancel()
	return m.DB.QueryRow(ctx, query, args...).Scan(&movie.ID, &movie.CreatedAt, &movie.Version)
}

func (m MovieModel) Get(id int64) (*Movie, error) {
	query := `
		SELECT id, created_at, title, year, runtime, genres, version 
		FROM movies
		WHERE id = $1
		`
	var movie Movie
	ctx, cancel := newQueryContext(3)
	defer cancel()
	err := m.DB.QueryRow(ctx, query, id).Scan(
		&movie.ID,
		&movie.CreatedAt,
		&movie.Title,
		&movie.Year,
		&movie.Runtime,
		&movie.Genres,
		&movie.Version,
	)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &movie, nil
}

func (m MovieModel) GetAll(title string, genres []string, filters Filters) ([]*Movie, error) {
	query := `
		SELECT id, created_at, title, year, runtime, genres, version
        FROM movies
        WHERE (TO_TSVECTOR('english', title) @@ PLAINTO_TSQUERY('english', $1) OR $1 = '')
        AND (genres @> $2 OR $2 = '{}')
        ORDER BY id
        `
	ctx, cancel := newQueryContext(3)
	defer cancel()
	rows, _ := m.DB.Query(ctx, query, title, genres)
	defer rows.Close()
	movies := make([]*Movie, 0)
	for rows.Next() {
		var movie Movie
		err := rows.Scan(
			&movie.ID,
			&movie.CreatedAt,
			&movie.Title,
			&movie.Year,
			&movie.Runtime,
			&movie.Genres,
			&movie.Version,
		)
		if err != nil {
			return nil, err
		}
		movies = append(movies, &movie)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return movies, nil
}

func (m MovieModel) Update(movie *Movie) error {
	query := `
		UPDATE movies 
		SET title = $1, year = $2, runtime = $3, genres = $4, version = version + 1
		WHERE id = $5 AND version = $6
		RETURNING version
		`
	args := []any{movie.Title, movie.Year, movie.Runtime, movie.Genres, movie.ID, movie.Version}
	ctx, cancel := newQueryContext(3)
	defer cancel()
	if err := m.DB.QueryRow(ctx, query, args...).Scan(&movie.Version); err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}
	return nil
}

func (m MovieModel) Delete(id int64) error {
	query := `
		DELETE FROM movies 
        WHERE id = $1
        `
	ctx, cancel := newQueryContext(3)
	defer cancel()
	status, err := m.DB.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if status.RowsAffected() == 0 {
		return ErrRecordNotFound
	}
	return nil
}
