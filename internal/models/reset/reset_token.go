package reset_models

import (
	"time"

	"github.com/jmoiron/sqlx"
)

// authid_model table
type ResetToken struct {
	ID        int64     `db:"id"`
	Hashed    string    `db:"hashed"`
	IsUsed    bool      `db:"is_used"`
	UserId    int64     `db:"user_id"`
	TTL       time.Time `db:"ttl"`
	ExpiresAt time.Time `db:"expires_at"`
	CreatedAt time.Time `db:"created_at"`
}

// id BIGSERIAL PRIMARY KEY,
//     hashed TEXT UNIQUE NOT NULL,
//     is_used BOOLEAN NOT NULL DEFAULT FALSE,
//     user_id BIGINT NULL, -- can be null for deleted users
//     ttl TIMESTAMPTZ DEFAULT NULL, -- so it can be recycled (1 year maybe)
//     expires_at TIMESTAMPTZ NOT NULL,
//     created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,

type ResetTokenDbService struct {
	DB *sqlx.DB
}

func NewResetTokenDbService(db *sqlx.DB) *ResetTokenDbService {
	return &ResetTokenDbService{
		DB: db,
	}
}

func (s *ResetTokenDbService) Create(userId int64, hashedToken string) (*ResetToken, error) {
	query := `
	INSERT INTO reset_token_model (user_id, hashed)
	VALUES ($1, $2)
	RETURNING id, hashed, is_used, user_id, ttl, expires_at, created_at
	`

	var token ResetToken
	if err := s.DB.Get(&token, query, userId, hashedToken); err != nil {
		return nil, err
	}

	return &token, nil
}

func (s *ResetTokenDbService) ReadByHashed(hashed string) (*ResetToken, error) {
	query := `
	SELECT id, hashed, is_used, user_id, ttl, expires_at, created_at
	FROM reset_token_model
	WHERE hashed = $1
	`

	var token ResetToken
	if err := s.DB.Get(&token, query, hashed); err != nil {
		return nil, err
	}
	return &token, nil
}

func (s *ResetTokenDbService) ReadAllByUserIdWithTx(tx *sqlx.Tx, userId int64) ([]ResetToken, error) {
	query := `
	SELECT id, hashed, is_used, user_id, ttl, expires_at, created_at
	FROM reset_token_model
	WHERE user_id = $1
	`

	var tokens []ResetToken
	// rows, err := service.DB.Query(query, ownerId)
	if err := tx.Select(&tokens, query, userId); err != nil {
		return nil, err
	}

	return tokens, nil
}

func (s *ResetTokenDbService) UpdateIsUsed(t *ResetToken, isUsed bool) error {
	query := `
	UPDATE reset_token_model
	SET is_used = $1
	WHERE id = $2
	RETURNING id, hashed, is_used, user_id, ttl, expires_at, created_at
	`

	if err := s.DB.Get(t, query, isUsed, t.ID); err != nil {
		return err
	}

	return nil
}

func (s *ResetTokenDbService) UpdateIsUsedWithTx(tx *sqlx.Tx, t *ResetToken, isUsed bool) error {
	query := `
	UPDATE reset_token_model
	SET is_used = $1
	WHERE id = $2
	RETURNING id, hashed, is_used, user_id, ttl, expires_at, created_at
	`

	if err := tx.Get(t, query, isUsed, t.ID); err != nil {
		return err
	}

	return nil
}

func (s *ResetTokenDbService) UpdateUserIdWithTx(tx *sqlx.Tx, t *ResetToken, userId int64) error {
	query := `
	UPDATE reset_token_model
	SET user_id = NULLIF($1, 0)
	WHERE id = $2
	RETURNING id, hashed, is_used, user_id, ttl, expires_at, created_at
	`

	if err := tx.Get(t, query, userId, t.ID); err != nil {
		return err
	}

	return nil
}
