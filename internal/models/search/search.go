package search_models

import (
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
)

type SearchResult struct {
	ID          int64          `db:"id"`
	Slug        string         `db:"slug"`
	Title       string         `db:"title"`
	Description sql.NullString `db:"description"`
	CreatedAt   time.Time      `db:"created_at"`
	ItemType    string         `db:"item_type"`
}

type SearchDbService struct {
	DB *sqlx.DB
}

func NewSearchDbService(db *sqlx.DB) *SearchDbService {
	return &SearchDbService{
		DB: db,
	}
}

func (s *SearchDbService) ReadPaginated(input string, page, perPage int) ([]*SearchResult, error) {
	offset := (page - 1) * perPage
	if offset < 1 {
		offset = 0 // safeguard
	}

	query := `
	SELECT id, slug, title, description, created_at, item_type
	FROM (
		SELECT
			b.id,
			b.slug,
			b.title,
			b.description,
			b.created_at,
			'board' AS item_type
		FROM board_model b
		INNER JOIN user_model u ON u.id = b.owner_id
		WHERE (
			b.title ILIKE '%' || $1 || '%'
			OR b.description ILIKE '%' || $1 || '%'
		)
		AND u.deactivated_at IS NULL
		AND NOT EXISTS (
			SELECT 1
			FROM email_ban_model eb
			WHERE eb.email = u.email
		)

		UNION ALL

		SELECT
			f.id,
			f.slug,
			f.title,
			f.description,
			f.created_at,
			'feedback' AS item_type
		FROM feedback_model f
		INNER JOIN board_model b ON b.id = f.board_id
		INNER JOIN user_model u ON u.id = b.owner_id
		WHERE (
			f.title ILIKE '%' || $1 || '%'
			OR f.description ILIKE '%' || $1 || '%'
		)
		AND f.expires_at IS NULL
		AND f.deleted_at IS NULL
		AND u.deactivated_at IS NULL
		AND NOT EXISTS (
			SELECT 1
			FROM email_ban_model eb
			WHERE eb.email = u.email
		)
	) AS results
	ORDER BY created_at DESC
	LIMIT $2
	OFFSET $3;
	`

	rows, err := s.DB.Queryx(query, input, perPage, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*SearchResult
	for rows.Next() {
		var r SearchResult
		if err := rows.StructScan(&r); err != nil {
			return nil, err
		}

		results = append(results, &r)
	}
	// Check for errors that occurred during iteration
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil

}

func (s *SearchDbService) ReadCount(input string) (int, error) {
	query := `
	SELECT COUNT(*)
	FROM (
		SELECT b.id
		FROM board_model b
		INNER JOIN user_model u ON u.id = b.owner_id
		WHERE (
			b.title ILIKE '%' || $1 || '%'
			OR b.description ILIKE '%' || $1 || '%'
		)
		AND u.deactivated_at IS NULL
		AND NOT EXISTS (
			SELECT 1
			FROM email_ban_model eb
			WHERE eb.email = u.email
		)

		UNION ALL

		SELECT f.id
		FROM feedback_model f
		INNER JOIN board_model b ON b.id = f.board_id
		INNER JOIN user_model u ON u.id = b.owner_id
		WHERE (
			f.title ILIKE '%' || $1 || '%'
			OR f.description ILIKE '%' || $1 || '%'
		)
		AND f.expires_at IS NULL
		AND f.deleted_at IS NULL
		AND u.deactivated_at IS NULL
		AND NOT EXISTS (
			SELECT 1
			FROM email_ban_model eb
			WHERE eb.email = u.email
		)
	) AS results;
	`
	var total int
	if err := s.DB.QueryRow(query, input).Scan(&total); err != nil {
		return 0, err
	}
	return total, nil

}
