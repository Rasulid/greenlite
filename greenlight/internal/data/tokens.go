package data

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base32"
	"time"

	"greenlight.rasulabduvaitov.net/internal/validator"
)


const (
	ScopeActiation = "activation"
)

type Token struct {
	PlainText string
	Hash []byte
	UserID int64
	Expiry time.Time
	Scope string
}


func generateToken(userID int64, ttl time.Duration, scope string) (*Token, error) {

	token := &Token{
		UserID: userID,
		Expiry: time.Now().Add(ttl),
		Scope: scope,
	}

	randomeBytes := make([]byte, 16)

	_, err := rand.Read(randomeBytes)
	if err != nil {
		return nil, err
	}

	token.PlainText = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomeBytes)

	hash := sha256.Sum256([]byte(token.PlainText))
	token.Hash = hash[:]


	return token, nil
}


func ValidateTokenPlainText(v *validator.Validator, tokenPlainText string)  {
	v.Check(tokenPlainText != "", "token", "must be provided")
	v.Check(len(tokenPlainText) == 26, "token", "must be 26 bytes long")
}


type TokenModel struct {
	DB *sql.DB
}


func (m *TokenModel) New(UserID int64, ttl time.Duration, scope string) (*Token, error) {

	token, err := generateToken(UserID, ttl, scope)
	if err != nil {
		return nil, err
	}

	err = m.Insert(token)

	return token, err

}


func (m *TokenModel) Insert(token *Token) error {

	query := `
	
			INSERT INTO tokens (hash, user_id, expiry, scope)
			VALUES ($1, $2, $3, $4)
	
	
	`

	args := []interface{}{token.Hash, token.UserID, token.Expiry, token.Scope}

	ctx , cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_,err := m.DB.ExecContext(ctx, query, args...)
	return err


}


func (m *TokenModel) DeleteAllForUser(UserID int64, scope string) error {

	query := `
			DELETE FROM tokens
			WHERE scope = $1, user_id = $2

	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, scope, UserID)

	return err

}