package email_ban_models

import (
	"time"

	"github.com/jmoiron/sqlx"
)

type EmailBanDbService struct {
	DB *sqlx.DB
}

func NewEmailBanDbService(db *sqlx.DB) *EmailBanDbService {
	return &EmailBanDbService{
		DB: db,
	}
}

// CREATE TABLE IF NOT EXISTS email_ban_model (
//     id BIGSERIAL PRIMARY KEY,
//     email TEXT UNIQUE NOT NULL,
//     reason TEXT NOT NULL,
//     created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
// );

type EmailBan struct {
	ID        int64     `db:"id"`
	Email     string    `db:"email"`
	Reason    string    `db:"reason"`
	CreatedAt time.Time `db:"created_at"`
}

func (s *EmailBanDbService) Create(email, reason string) (*EmailBan, error) {
	query := `
	INSERT INTO email_ban_model (email, reason)
	VALUES ($1, $2)
	RETURNING id, email, reason, created_at
	`
	var b EmailBan
	if err := s.DB.Get(&b, query, email, reason); err != nil {
		return nil, err
	}

	return &b, nil
}

func (s *EmailBanDbService) ReadByEmail(email string) (*EmailBan, error) {
	query := `
	SELECT id, email, reason, created_at
	FROM email_ban_model
	WHERE email = $1
	`
	var b EmailBan
	if err := s.DB.Get(&b, query, email); err != nil {
		return nil, err
	}
	return &b, nil
}
