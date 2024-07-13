package data

import (
	"errors"
	"github.com/M0hammadUsman/greenlight/internal/validator"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
	"time"
)

var (
	ErrDuplicateEmail = errors.New("duplicate email")
)

type User struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  password  `json:"-"`
	Activated bool      `json:"activated"`
	Version   int       `json:"-"`
}

type password struct {
	plainText *string
	hash      []byte
}

func (p *password) Set(plainTextPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plainTextPassword), 12)
	if err != nil {
		return err
	}
	p.hash = hash
	p.plainText = &plainTextPassword
	return nil
}

func (p *password) Matches(plainTextPassword string) (bool, error) {
	if err := bcrypt.CompareHashAndPassword(p.hash, []byte(plainTextPassword)); err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, nil
		}
	}
	return true, nil
}

func ValidateEmail(v *validator.Validator, email string) {
	v.Check(email != "", "email", "email must be provided")
	v.Check(validator.ValidEmail(email), "email", "must be a valid email address")
}

func validatePasswordPlainText(v *validator.Validator, password string) {
	v.Check(password != "", "password", "must be provided")
	v.Check(len(password) >= 8, "password", "must be at least 8 bytes long")
	v.Check(len(password) <= 72, "password", "must not be more than 500 bytes long")
}

func ValidateUser(v *validator.Validator, user *User) {
	v.Check(user.Name != "", "name", "must be provided")
	v.Check(len(user.Name) <= 500, "name", "must not be more than 500 bytes long")
	ValidateEmail(v, user.Email)
	if user.Password.plainText != nil {
		validatePasswordPlainText(v, *user.Password.plainText)
	}
	/*If the password hash is ever nil, this will be due to a logic error in our codebase (probably because we forgot
	to set a password for the user). It's a useful sanity check to include here, but it's not a problem with the data
	provided by the client. So rather than adding an error to the validation map we raise a panic instead.*/
	if user.Password.hash == nil {
		panic("missing password hash for user")
	}
}

type UserModel struct {
	DB *pgxpool.Pool
}

func (m UserModel) Insert(user *User) error {
	query := `
		INSERT INTO users (name, email, password, activated)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, version
		`
	args := []any{user.Name, user.Email, user.Password.hash, user.Activated}
	ctx, cancel := newQueryContext(3)
	defer cancel()
	if err := m.DB.QueryRow(ctx, query, args...).Scan(&user.ID, &user.CreatedAt, &user.Version); err != nil {
		var pgErr *pgconn.PgError
		switch {
		case errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == "users_email_key":
			return ErrDuplicateEmail
		default:
			return err
		}
	}
	return nil
}

func (m UserModel) GetByEmail(email string) (*User, error) {
	query := `
		SELECT id, created_at, name, email, password, activated, version 
		FROM users
		WHERE email = $1
		`
	var user User
	ctx, cancel := newQueryContext(3)
	defer cancel()
	if err := m.DB.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.Name,
		&user.Email,
		&user.Password,
		&user.Activated,
		&user.Version,
	); err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &user, nil
}

func (m UserModel) UpdateUser(user *User) error {
	query := `
		UPDATE users 
		SET name = $1, email = $2, password = $3, activated = $4, version = version + 1
		WHERE id = $5 AND version = $6
		RETURNING version
		`
	args := []any{user.Name, user.Email, user.Password, user.Activated, user.ID, user.Version}
	ctx, cancel := newQueryContext(3)
	defer cancel()
	var version int
	if err := m.DB.QueryRow(ctx, query, args...).Scan(&version); err != nil {
		var pgErr *pgconn.PgError
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return ErrEditConflict
		case errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == "users_email_key":
			return ErrDuplicateEmail
		default:
			return err
		}
	}
	return nil
}
