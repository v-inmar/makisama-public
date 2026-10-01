package image_models

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

// CREATE TABLE IF NOT EXISTS board_image_model (
//     id BIGSERIAL PRIMARY KEY,
//     board_id BIGINT UNIQUE NOT NULL,
//     image_id BIGINT UNIQUE NOT NULL,
//     created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
// );

type BoardImage struct {
	ID        int64     `db:"id"`
	BoardId   int64     `db:"board_id"`
	ImageId   int64     `db:"image_id"`
	CreatedAt time.Time `db:"created_at"`
}

type BoardImageDbService struct {
	DB *sqlx.DB
}

func NewBoardImageDbService(db *sqlx.DB) *BoardImageDbService {
	return &BoardImageDbService{
		DB: db,
	}
}

func (s *BoardImageDbService) CreateWithTx(tx *sqlx.Tx, boardId, imageId int64) (*BoardImage, error) {
	query := `
	INSERT INTO board_image_model (board_id, image_id)
	VALUES ($1, $2)
	RETURNING id, board_id, image_id, created_at
	`
	var boardImage BoardImage
	if err := tx.Get(&boardImage, query, boardId, imageId); err != nil {
		return nil, err
	}

	return &boardImage, nil
}

func (s *BoardImageDbService) ReadByBoardId(boardId int64) (*BoardImage, error) {
	query := `
	SELECT id, board_id, image_id, created_at
	FROM board_image_model
	WHERE board_id = $1
	`

	var boardImage BoardImage
	if err := s.DB.Get(&boardImage, query, boardId); err != nil {
		return nil, err
	}

	return &boardImage, nil
}

func (s *BoardImageDbService) ReadByBoardIdWithTx(tx *sqlx.Tx, boardId int64) (*BoardImage, error) {
	query := `
	SELECT id, board_id, image_id, created_at
	FROM board_image_model
	WHERE board_id = $1
	`

	var boardImage BoardImage
	if err := tx.Get(&boardImage, query, boardId); err != nil {
		return nil, err
	}

	return &boardImage, nil
}

func (s *BoardImageDbService) UpdateImageWithTx(tx *sqlx.Tx, bim *BoardImage, imgId int64) error {
	query := `
	UPDATE board_image_model
	SET image_id = $1
	WHERE id = $2
	RETURNING id, board_id, image_id, created_at
	`

	if err := tx.Get(bim, query, imgId, bim.ID); err != nil {
		return err
	}

	return nil

}

func (s *BoardImageDbService) DeleteWithTx(tx *sqlx.Tx, bim *BoardImage) error {
	query := `
	DELETE FROM board_image_model
	WHERE id = $1
	`

	result, err := tx.Exec(query, bim.ID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no board found with id %d", bim.ID)
	}

	return nil
}
