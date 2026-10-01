package board_models

import (
	"app/internal/constants"
	randoms_utils "app/internal/utilities/randoms"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/gosimple/slug"
	"github.com/jmoiron/sqlx"
)

// table name board_model
type Board struct {
	ID          int64          `db:"id"`
	CreatedAt   time.Time      `db:"created_at"`
	UpdatedAt   sql.NullTime   `db:"updated_at"`
	Title       string         `db:"title"`
	Website     sql.NullString `db:"website"`
	Description sql.NullString `db:"description"`
	Slug        string         `db:"slug"`
	OwnerID     int64          `db:"owner_id"`
}

type BoardDbService struct {
	DB *sqlx.DB
}

func NewBoardDbService(db *sqlx.DB) *BoardDbService {
	return &BoardDbService{
		DB: db,
	}
}

func (service *BoardDbService) CreateWithTx(tx *sqlx.Tx, title, website, description string, userId int64) (*Board, error) {
	var boardSlug string
	s := slug.Make(title)
	for range 5 {

		// prevents slug using endpoints that are used by this app
		// intend to just make it as random as possible without losing actual title
		if _, exist := constants.ReservedWords[strings.ToLower(s)]; exist {
			appendString := strconv.FormatInt(randoms_utils.GenRandNumBetween(111111, math.MaxInt64), 10)
			s = s + "-" + appendString // i.e. dashboard-123456
		}
		// if slices.Contains(constants.ResrvedWords, strings.ToLower(s)) {
		// 	appendString := strconv.FormatInt(randoms_utils.GenRandNumBetween(111111, math.MaxInt64), 10)
		// 	s = s + "-" + appendString // i.e. dashboard-123456
		// }

		var exists bool
		slugQuery := `
			SELECT EXISTS (
				SELECT 1
				FROM board_model
				WHERE slug = $1
			); 
		`
		if err := tx.QueryRow(slugQuery, s).Scan(&exists); err != nil {
			continue
		}

		if !exists {
			boardSlug = s
			break
		}

		// just add more random characters
		addedString := strconv.FormatInt(randoms_utils.GenRandNumBetween(111111, math.MaxInt64), 10)
		s = s + "-" + addedString
	}
	if boardSlug == "" {
		return nil, errors.New("failed to generate slug")
	}

	var board Board

	insertQuery := `
	INSERT INTO board_model (title, website, description, slug, owner_id)
	VALUES ($1, NULLIF($2, ''), NULLIF($3, ''), $4, $5)
	RETURNING id, created_at, updated_at, title, website, description, slug, owner_id
	`
	if err := tx.Get(&board, insertQuery, title, website, description, boardSlug, userId); err != nil {
		return nil, err
	}

	return &board, nil

}

func (service *BoardDbService) Create(title, website, description string, userId int64) (*Board, error) {
	var boardSlug string
	s := slug.Make(title)
	for range 5 {

		// prevents slug using endpoints that are used by this app
		// intend to just make it as random as possible without losing actual title
		if _, exist := constants.ReservedWords[strings.ToLower(s)]; exist {
			appendString := strconv.FormatInt(randoms_utils.GenRandNumBetween(111111, math.MaxInt64), 10)
			s = s + "-" + appendString // i.e. dashboard-123456
		}
		// if slices.Contains(constants.ResrvedWords, strings.ToLower(s)) {
		// 	appendString := strconv.FormatInt(randoms_utils.GenRandNumBetween(111111, math.MaxInt64), 10)
		// 	s = s + "-" + appendString // i.e. dashboard-123456
		// }

		var exists bool
		slugQuery := `
			SELECT EXISTS (
				SELECT 1
				FROM board_model
				WHERE slug = $1
			); 
		`
		if err := service.DB.QueryRow(slugQuery, s).Scan(&exists); err != nil {
			continue
		}

		if !exists {
			boardSlug = s
			break
		}

		// just add more random characters
		addedString := strconv.FormatInt(randoms_utils.GenRandNumBetween(111111, math.MaxInt64), 10)
		s = s + "-" + addedString
	}
	if boardSlug == "" {
		return nil, errors.New("failed to generate slug")
	}

	var board Board

	insertQuery := `
	INSERT INTO board_model (title, website, description, slug, owner_id)
	VALUES ($1, NULLIF($2, ''), NULLIF($3, ''), $4, $5)
	RETURNING id, created_at, updated_at, title, website, description, slug, owner_id
	`
	if err := service.DB.Get(&board, insertQuery, title, website, description, boardSlug, userId); err != nil {
		return nil, err
	}

	return &board, nil

}

// CreateNewBoard creates a new board with the specified title, website, and description,
// and associates it with the given user.
//
// Parameters:
//   - title: The title of the board. Used to generate a unique slug.
//   - website: The optional website URL of the board. Empty strings are stored as NULL.
//   - description: The optional description of the board. Empty strings are stored as NULL.
//   - user_id: The ID of the user who will own the board.
//
// Returns:
//   - *Board: A pointer to the newly created Board struct, including its ID, slug,
//     and timestamps.
//   - error: An error if slug generation fails, the title is invalid, or the database
//     insert fails.
//
// This function generates a unique slug from the board title, avoiding conflicts with
// reserved words and existing slugs in the database. It attempts multiple times to
// ensure uniqueness by appending random numbers if necessary. After a unique slug is
// generated, the board is inserted into the `board_model` table and returned.
// func (service *BoardDbService) CreateNewBoard(title, website, description string, user_id int64) (*Board, error) {
// 	var boardSlug string
// 	s := slug.Make(title)
// 	for range 5 {

// 		// prevents slug using endpoints that are used by this app
// 		// intend to just make it as random as possible without losing actual title
// 		if slices.Contains(constants.ResrvedWords, strings.ToLower(s)) {
// 			appendString := strconv.FormatInt(randoms_utils.GenRandNumBetween(111111, math.MaxInt64), 10)
// 			s = s + "-" + appendString // i.e. dashboard-123456
// 		}

// 		var exists bool
// 		slugQuery := `
// 			SELECT EXISTS (
// 				SELECT 1
// 				FROM board_model
// 				WHERE slug = $1
// 			);
// 		`
// 		if err := service.DB.QueryRow(slugQuery, s).Scan(&exists); err != nil {
// 			continue
// 		}

// 		if !exists {
// 			boardSlug = s
// 			break
// 		}

// 		// just add more random characters
// 		addedString := strconv.FormatInt(randoms_utils.GenRandNumBetween(111111, math.MaxInt64), 10)
// 		s = s + "-" + addedString
// 	}
// 	if boardSlug == "" {
// 		return nil, errors.New("failed to generate slug")
// 	}

// 	var board Board

// 	insertQuery := `
// 	INSERT INTO board_model (title, website, description, slug, owner_id)
// 	VALUES ($1, NULLIF($2, ''), NULLIF($3, ''), $4, $5)
// 	RETURNING id, created_at, updated_at, title, website, description, slug, owner_id
// 	`
// 	if err := service.DB.Get(&board, insertQuery, title, website, description, boardSlug, user_id); err != nil {
// 		return nil, err
// 	}

// 	return &board, nil

// }

// ReadCountActiveBoardsByUserId returns the total number of active boards
// owned by the specified user.
//
// Parameters:
//   - userId: The unique identifier of the board owner.
//
// Returns:
//   - int: The count of active boards owned by the user.
//   - error: An error if the query execution fails.
//
// This function queries the `board_model` table for boards belonging to the
// specified user where `deleted_at` is NULL, ensuring that only active boards
// are counted.
func (service *BoardDbService) ReadCountActiveBoardsByUserId(userId int64) (int, error) {
	var total int

	query := `
		SELECT COUNT(*)
		FROM board_model
		WHERE owner_id = $1
	`

	if err := service.DB.QueryRow(query, userId).Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}

// ReadPaginatedActiveBoardsByUserId retrieves a paginated list of active boards
// owned by the specified user.
//
// Parameters:
//   - userId: The unique identifier of the board owner.
//   - page: The page number to retrieve (1-based).
//   - perPage: The maximum number of boards to return per page.
//
// Returns:
//   - []Board: A slice of active boards owned by the user, ordered by creation
//     date in descending order.
//   - error: An error if the query execution or row scanning fails.
//
// This function queries the `board_model` table for boards belonging to the
// specified user where `deleted_at` is NULL, ensuring that only active boards
// are returned. Results are ordered by `created_at` in descending order and
// paginated using LIMIT and OFFSET.
func (service *BoardDbService) ReadPaginatedActiveBoardsByUserId(userId int64, page, perPage int) ([]Board, error) {
	offset := (page - 1) * perPage
	if offset < 1 {
		offset = 0 // safeguard
	}

	query := `
	SELECT id, created_at, updated_at, title, website, description, slug, owner_id
	FROM board_model
	WHERE owner_id = $1
	ORDER BY created_at DESC
	LIMIT $2 OFFSET $3
	`

	rows, err := service.DB.Query(query, userId, perPage, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var boards []Board

	for rows.Next() {
		var b Board
		err := rows.Scan(
			&b.ID,
			&b.CreatedAt,
			&b.UpdatedAt,
			&b.Title,
			&b.Website,
			&b.Description,
			&b.Slug,
			&b.OwnerID,
		)

		if err != nil {
			return nil, err
		}

		boards = append(boards, b)
	}

	return boards, nil
}

// ReadActiveBoardById retrieves a board by its ID if it is active (i.e., not deleted).
//
// Parameters:
//   - id: The unique identifier of the board to retrieve.
//
// Returns:
//   - *Board: A pointer to the Board struct if found and active.
//   - error: An error if the board is not found or a database error occurs.
//
// This function queries the `board_model` table for a board with the given ID
// where `deleted_at` is NULL, ensuring that only active boards are returned.
// func (service *BoardDbService) ReadActiveBoardById(id int64) (*Board, error) {
// 	query := `
// 	SELECT id, deleted_at, created_at, title, website, description, slug, owner_id
// 	FROM board_model
// 	WHERE id=$1
// 	AND deleted_at IS NULL
// 	`
// 	var board Board
// 	err := service.DB.Get(&board, query, id)

// 	if err != nil {
// 		return nil, err
// 	}

// 	return &board, nil
// }

func (service *BoardDbService) ReadAllByOwnerId(ownerId int64) ([]Board, error) {
	query := `
	SELECT id, created_at, title, website, description, slug, owner_id
	FROM board_model
	WHERE owner_id=$1
	`
	var boards []Board
	// rows, err := service.DB.Query(query, ownerId)
	if err := service.DB.Select(&boards, query, ownerId); err != nil {
		return nil, err
	}

	return boards, nil

}

func (s *BoardDbService) ReadById(id int64) (*Board, error) {
	query := `
	SELECT id, created_at, updated_at, title, website, description, slug, owner_id
	FROM board_model
	WHERE id=$1 
	`

	var b Board
	if err := s.DB.Get(&b, query, id); err != nil {
		return nil, err
	}

	return &b, nil
}

func (s *BoardDbService) ReadBySlug(slug string) (*Board, error) {
	query := `
	SELECT id, created_at, updated_at, title, website, description, slug, owner_id
	FROM board_model
	WHERE slug=$1 
	`

	var b Board
	if err := s.DB.Get(&b, query, slug); err != nil {
		return nil, err
	}

	return &b, nil
}

func (s *BoardDbService) Update(board *Board, title, website, description string) error {
	query := `
	UPDATE board_model
	SET title = $1, website = NULLIF($2, ''), description = NULLIF($3, ''), updated_at = NOW()
	WHERE id = $4
	RETURNING id, created_at, updated_at, title, website, description, slug, owner_id
	`
	if err := s.DB.Get(board, query, title, website, description, board.ID); err != nil {
		return err
	}

	return nil
}

func (s *BoardDbService) UpdateWithTx(tx *sqlx.Tx, b *Board, title, website, description string) error {
	query := `
	UPDATE board_model
	SET title = $1, website = NULLIF($2, ''), description = NULLIF($3, ''), updated_at = NOW()
	WHERE id = $4
	RETURNING id, created_at, updated_at, title, website, description, slug, owner_id
	`

	if err := tx.Get(b, query, title, website, description, b.ID); err != nil {
		return err
	}

	return nil

}

func (service *BoardDbService) DeleteBoardWithTx(tx *sqlx.Tx, board *Board) error {
	query := `
	DELETE FROM board_model
	WHERE id = $1
	`

	res, err := tx.Exec(query, board.ID)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("no board found with id %d", board.ID)
	}

	return nil
}

func (s *BoardDbService) Delete(board *Board) error {
	query := `
	DELETE FROM board_model
	WHERE id = $1
	`

	result, err := s.DB.Exec(query, board.ID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no board found with id %d", board.ID)
	}

	return nil
}

func (s *BoardDbService) DeleteWithTx(tx *sqlx.Tx, board *Board) error {
	query := `
	DELETE FROM board_model
	WHERE id = $1
	`

	result, err := tx.Exec(query, board.ID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no board found with id %d", board.ID)
	}

	return nil
}
