package data

import (
	"database/sql"
	"errors"
)

var (
	ErrorRecordNotFound = errors.New("record Not Found")
	ErrorEditConflict = errors.New("edit conflict")
)

type Models struct {
	Movies MovieModel
	Users UserModel
	Token TokenModel
}

func NewMovies(db *sql.DB) Models {
	return Models{
		Movies: MovieModel{DB: db},
		Users: UserModel{DB: db},
		Token: TokenModel{DB: db},
	}
}
