package confirm_email_models

import (
	"time"

	"github.com/jmoiron/sqlx"
)

// CREATE TABLE IF NOT EXISTS user_confirm_model (
//
//	id BIGSERIAL PRIMARY KEY,
//	user_id BIGINT NOT NULL,
//	confirm_id BIGINT NOT NULL,
//	created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
//
// )

type UserConfirm struct {
	ID        int64     `db:"id"`
	UserId    int64     `db:"user_id"`
	ConfirmId int64     `db:"confirm_id"`
	CreatedAt time.Time `db:"created_at"`
}

type UserConfirmDbService struct {
	DB *sqlx.DB
}

func NewUserConfirmDbService(db *sqlx.DB) *UserConfirmDbService {
	return &UserConfirmDbService{
		DB: db,
	}
}

func (s *UserConfirmDbService) CreateWithTx(tx *sqlx.Tx, userId, confirmId int64) (*UserConfirm, error) {
	query := `
	INSERT INTO user_confirm_model (user_id, confirm_id)
	VALUES ($1, $2)
	RETURNING id, user_id, confirm_id, created_at
	`

	var m UserConfirm
	if err := tx.Get(&m, query, userId, confirmId); err != nil {
		return nil, err
	}
	return &m, nil
}

func (s *UserConfirmDbService) ReadByConfirmID(confirmId int64) (*UserConfirm, error) {
	query := `
	SELECT id, user_id, confirm_id, created_at
	FROM user_confirm_model
	WHERE confirm_id=$1
	`
	var m UserConfirm
	if err := s.DB.Get(&m, query, confirmId); err != nil {
		return nil, err
	}
	return &m, nil
}

func (s *UserConfirmDbService) ReadByUserID(userId int64) (*UserConfirm, error) {
	query := `
	SELECT id, user_id, confirm_id, created_at
	FROM user_confirm_model
	WHERE user_id=$1
	`
	var m UserConfirm
	if err := s.DB.Get(&m, query, userId); err != nil {
		return nil, err
	}
	return &m, nil
}

func (s *UserConfirmDbService) ReadByUserIDWithTx(tx *sqlx.Tx, userId int64) (*UserConfirm, error) {
	query := `
	SELECT id, user_id, confirm_id, created_at
	FROM user_confirm_model
	WHERE user_id=$1
	`
	var m UserConfirm
	if err := tx.Get(&m, query, userId); err != nil {
		return nil, err
	}
	return &m, nil
}

func (s *UserConfirmDbService) UpdateConfirmIDWithTx(tx *sqlx.Tx, uc *UserConfirm, cid int64) error {
	query := `
	UPDATE user_confirm_model
	SET confirm_id = $1
	WHERE id = $2
	RETURNING id, user_id, confirm_id, created_at
	`
	if err := tx.Get(uc, query, cid, uc.ID); err != nil {
		return err
	}

	return nil
}
