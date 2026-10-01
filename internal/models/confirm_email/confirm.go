package confirm_email_models

import (
	"time"

	"github.com/jmoiron/sqlx"
)

// CREATE TABLE IF NOT EXISTS confirm_token_model (
//     id BIGSERIAL PRIMARY KEY,
//     hashed TEXT UNIQUE NOT NULL,
//     expires_at TIMESTAMPTZ NOT NULL DEFAULT (NOW() + INTERVAL '24 hour'),
//     created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
// );

type ConfirmToken struct {
	ID        int64     `db:"id"`
	Hashed    string    `db:"hashed"`
	ExpiresAt time.Time `db:"expires_at"`
	CreatedAt time.Time `db:"created_at"`
}

type ConfirmTokenDbSerivce struct {
	DB *sqlx.DB
}

func NewConfirmTokenDbSerivce(db *sqlx.DB) *ConfirmTokenDbSerivce {
	return &ConfirmTokenDbSerivce{
		DB: db,
	}
}

func (s *ConfirmTokenDbSerivce) CreateWithTx(tx *sqlx.Tx, hashed string) (*ConfirmToken, error) {
	query := `
	INSERT INTO confirm_token_model (hashed)
	VALUES ($1)
	RETURNING id, hashed, expires_at, created_at
	`

	var c ConfirmToken
	if err := tx.Get(&c, query, hashed); err != nil {
		return nil, err
	}
	return &c, nil
}

func (s *ConfirmTokenDbSerivce) ReadByHashed(hashed string) (*ConfirmToken, error) {
	query := `
	SELECT id, hashed, expires_at, created_at
	FROM confirm_token_model
	WHERE hashed=$1
	`

	var c ConfirmToken
	if err := s.DB.Get(&c, query, hashed); err != nil {
		return nil, err
	}
	return &c, nil
}
