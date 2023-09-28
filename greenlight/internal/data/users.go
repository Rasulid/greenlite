package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
	"greenlight.rasulabduvaitov.net/internal/validator"
)


var (
	ErrorDublicateEmail = errors.New("dublicate email")  
)


type UserModel struct {
	DB *sql.DB
}


type User struct {
	ID string `json:"id"`
	CreatedAt time.Time `json:"created"`
	Name string `json:"name"`
	Email string `json:"email"`
	Password password `json:"-"`
	Activate bool `json:"activate"`
	Version int `json:"version"`

}

type password struct {
	plaintext *string
	hash []byte
}


func (p *password) Set(plaintextPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintextPassword), 12)
	if err != nil {
		return err
	}

	p.plaintext = &plaintextPassword
	p.hash = hash

	return nil
}

func (p *password) Matches(plaintextPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(p.hash, []byte(plaintextPassword))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, err
		default:
			return false, err
		}
	}

	return true, err
}

func ValideteEmail(v *validator.Validator, email string){
	v.Check(email != "", "email", "must be provided")
	v.Check(validator.Matches(email, validator.EmailRx), "email", "must be valid email address")
}

func ValidetePasswordPlaintext(v *validator.Validator, password string) {
	v.Check(password != "", "password", "must be provided")
	v.Check(len(password) >= 8, "password", "must be at least 8 bites long")
	v.Check(len(password) <= 72, "password", "must not be more then 72, bytes long")
}

func ValidateUser(v *validator.Validator, user *User) {
	v.Check(user.Name != "", "name", "must be providet")
	v.Check(len(user.Name) <= 500, "name", "must not be more then 500 bytes long")
	ValideteEmail(v, user.Email)

	if user.Password.plaintext != nil {
		ValidetePasswordPlaintext(v, *user.Password.plaintext)
	}


	if user.Password.hash == nil {
		panic("missing password hash for user")


	}
}


func (m UserModel) Insert(user *User) error {
	query := `
				INSERT INTO users (name, email, password_hash, activated)
				VALUES ($1, $2, $3, $4)
				RETURNING id, created_at, version

	`

	args := []interface{}{user.Name, user.Email, string(user.Password.hash), user.Activate}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()


	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&user.ID, &user.CreatedAt, &user.Version)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code.Class() == "23" {
				return ErrorDublicateEmail
			}
		}
		switch{
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key" `:
			return ErrorDublicateEmail
		default:
			return err
		}
	}

	return nil
}

func (m *UserModel) GetByEmail(email string) (*User, error) {
	query := `
			SELECT id, created_at, name, email, password_hash, activate, version
			FROM users 
			WHERE email == $1
	
	`

	var user User

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()


	err := m.DB.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.Name,
		&user.Email,
		&user.Password.hash,
		&user.Activate,
		&user.Version,
	)
	if err != nil {
		switch{
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrorRecordNotFound
		default:
			return nil, err
		}
	}

	return &user, nil
}


func (m *UserModel) Update(user *User) error {

	query := `
			UPDATE users
			SET name = $1, email = $2, password_hash = $3, activated = $4, version = version + 1
			WHERE id = $5 AND version  = $6
			RETURNING version
	`

	args := []interface{}{
		user.Name,
		user.Email,
		user.Password.hash,
		user.Activate,
		user.ID,
		user.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&user.Version)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrorDublicateEmail

		default:
			return err
		}
	}

	return nil

}