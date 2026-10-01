package company_models

import (
	"time"

	"github.com/jmoiron/sqlx"
)

// CREATE TABLE IF NOT EXISTS contact_model (
//     id BIGSERIAL PRIMARY KEY,
//     firstname TEXT NOT NULL,
//     lastname TEXT NOT NULL,
//     email TEXT NOT NULL,
//     message TEXT DEFAULT NULL,
//     created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
// );

type Contact struct {
	ID        int64     `db:"id"`
	Firstname string    `db:"firstname"`
	Lastname  string    `db:"lastname"`
	Email     string    `db:"email"`
	Message   string    `db:"message"`
	CreatedAt time.Time `db:"created_at"`
}

type ContactDbService struct {
	DB *sqlx.DB
}

func NewContactDbService(db *sqlx.DB) *ContactDbService {
	return &ContactDbService{
		DB: db,
	}
}

func (s *ContactDbService) Create(firstname, lastname, email, message string) (*Contact, error) {
	query := `
	INSERT INTO contact_model(firstname, lastname, email, message)
	VALUES ($1, $2, $3, $4)
	RETURNING id, firstname, lastname, email, message, created_at
	`
	var contact Contact

	if err := s.DB.Get(&contact, query, firstname, lastname, email, message); err != nil {
		return nil, err
	}

	return &contact, nil
}
