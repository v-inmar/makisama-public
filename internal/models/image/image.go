package image_models

import (
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
)

// CREATE TABLE IF NOT EXISTS image_model (
//     id BIGSERIAL PRIMARY KEY,
//     filename TEXT UNIQUE NOT NULL,
//     extension TEXT NOT NULL,
//     ttl TIMESTAMPTZ DEFAULT NULL,
//     created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
// );

// image_model
type Image struct {
	ID        int64        `db:"id"`
	Filename  string       `db:"filename"`
	Extension string       `db:"extension"`
	TTL       sql.NullTime `db:"ttl"`
	CreatedAt time.Time    `db:"created_at"`
}

type ImageDbService struct {
	DB *sqlx.DB
}

func NewImageDbService(db *sqlx.DB) *ImageDbService {
	return &ImageDbService{
		DB: db,
	}
}

func (s *ImageDbService) CreateWithTx(tx *sqlx.Tx, filename, ext string) (*Image, error) {
	query := `
	INSERT INTO image_model (filename, extension)
	VALUES ($1, $2)
	RETURNING id, filename, extension, ttl, created_at
	`

	var image Image
	if err := tx.Get(&image, query, filename, ext); err != nil {
		return nil, err
	}

	return &image, nil

}

func (s *ImageDbService) ReadByFilename(filename string) (*Image, error) {
	query := `
	SELECT id, filename, extension, ttl, created_at
	FROM image_model
	WHERE filename = $1
	`

	var i Image
	if err := s.DB.Get(&i, query, filename); err != nil {
		return nil, err
	}

	return &i, nil
}

func (s *ImageDbService) ReadById(id int64) (*Image, error) {
	query := `
	SELECT id, filename, extension, ttl, created_at
	FROM image_model
	WHERE id = $1
	`

	var i Image
	if err := s.DB.Get(&i, query, id); err != nil {
		return nil, err
	}

	return &i, nil
}

func (s *ImageDbService) ReadByIdWithTx(tx *sqlx.Tx, id int64) (*Image, error) {
	query := `
	SELECT id, filename, extension, ttl, created_at
	FROM image_model
	WHERE id = $1
	`

	var i Image
	if err := tx.Get(&i, query, id); err != nil {
		return nil, err
	}

	return &i, nil
}

func (s *ImageDbService) UpdateTTLWithTx(tx *sqlx.Tx, i *Image, ttl *time.Time) error {
	query := `
	UPDATE image_model
	SET ttl = $1
	WHERE id = $2
	RETURNING id, filename, extension, ttl, created_at
	`

	if err := tx.Get(i, query, ttl, i.ID); err != nil {
		return err
	}

	return nil
}
