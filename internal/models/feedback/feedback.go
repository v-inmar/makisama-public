package feedback_models

import (
	"app/internal/constants"
	date_utils "app/internal/utilities/dates"
	randoms_utils "app/internal/utilities/randoms"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

// feedback_model
type Feedback struct {
	ID          int64          `db:"id"`
	Title       string         `db:"title"`
	Description sql.NullString `db:"description"`
	Slug        string         `db:"slug"`
	CreatedAt   time.Time      `db:"created_at"`
	ExpiresAt   sql.NullTime   `db:"expires_at"`
	BoardID     int64          `db:"board_id"`
	OwnerID     sql.NullInt64  `db:"owner_id"`
	Email       sql.NullString `db:"email"`
	Scope       sql.NullString `db:"scope"`
	Resolution  sql.NullString `db:"resolution"`
	UpdatedAt   sql.NullTime   `db:"updated_at"`
	DeletedAt   sql.NullTime   `db:"deleted_at"`
}

type FeedbackDbService struct {
	DB *sqlx.DB
}

func NewFeedbackDbService(db *sqlx.DB) *FeedbackDbService {
	return &FeedbackDbService{
		DB: db,
	}
}

func (s *FeedbackDbService) Create(title, description, email string, board_id, owner_id int64) (*Feedback, error) {
	var slug string
	for range 5 {
		feedbackSlug, err := randoms_utils.GenerateFeedbackSlugValue()
		if err != nil {
			continue
		}

		slugQuery := `SELECT EXISTS ( SELECT 1 FROM feedback_model WHERE slug = $1);`

		var exists bool
		if err := s.DB.QueryRow(slugQuery, feedbackSlug).Scan(&exists); err != nil {
			continue
		}

		if !exists {
			slug = feedbackSlug
			break
		}

	}

	if slug == "" {
		return nil, errors.New("failed to generate slug")
	}

	expires_at := date_utils.GetMidnightOfSeventhDay()

	var feedback Feedback

	insertQuery := `
	INSERT INTO feedback_model (title, description, slug, expires_at, board_id, owner_id, email)
	VALUES ($1, NULLIF($2, ''), $3, $4, $5, NULLIF($6, 0), NULLIF($7, ''))
	RETURNING id, title, description, slug, expires_at, board_id, owner_id, email, scope, resolution, updated_at, deleted_at
	`
	if err := s.DB.Get(&feedback, insertQuery, title, description, slug, expires_at, board_id, owner_id, email); err != nil {
		return nil, err
	}

	return &feedback, nil
}

// CreateNewFeedback creates a new feedback entry for a given board and owner.
//
// The function generates a unique slug for the feedback using a secure random string. It attempts
// up to 5 times to generate a slug that does not already exist in the database. If all attempts
// fail, it returns an error.
//
// The feedback's expiration date is set to the midnight of the seventh day from the current date.
//
// During insertion:
//   - Empty description strings are stored as NULL.
//   - A zero value for owner_id is stored as NULL.
//   - Empty email strings are stored as NULL.
//
// Parameters:
//   - title: title of the feedback
//   - description: optional description of the feedback
//   - email: optional email of the feedback owner
//   - board_id: ID of the board this feedback belongs to
//   - owner_id: ID of the feedback owner
//
// Returns:
//   - *Feedback: the newly created feedback record, including its generated slug and timestamps
//   - error: database error, slug generation failure, or other insertion error
// func (service *FeedbackDbService) CreateNewFeedback(title, description, email string, board_id, owner_id int64) (*Feedback, error) {
// 	var slug string
// 	for range 5 {
// 		feedbackSlug, err := randoms_utils.SecureRandString(16, 32)
// 		if err != nil {
// 			continue
// 		}

// 		slugQuery := `SELECT EXISTS ( SELECT 1 FROM feedback_model WHERE slug = $1);`

// 		var exists bool
// 		if err := service.DB.QueryRow(slugQuery, feedbackSlug).Scan(&exists); err != nil {
// 			continue
// 		}

// 		if !exists {
// 			slug = feedbackSlug
// 			break
// 		}

// 	}

// 	if slug == "" {
// 		return nil, errors.New("failed to generate slug")
// 	}

// 	expires_at := date_utils.GetMidnightOfSeventhDay()

// 	var feedback Feedback

// 	insertQuery := `
// 	INSERT INTO feedback_model (title, description, slug, expires_at, board_id, owner_id, email)
// 	VALUES ($1, NULLIF($2, ''), $3, $4, $5, NULLIF($6, 0), NULLIF($7, ''))
// 	RETURNING id, title, description, slug, expires_at, board_id, owner_id, email, scope, resolution, updated_at, deleted_at
// 	`
// 	if err := service.DB.Get(&feedback, insertQuery, title, description, slug, expires_at, board_id, owner_id, email); err != nil {
// 		return nil, err
// 	}
// 	return &feedback, nil
// }

// TODO: Create docstring
func (service *FeedbackDbService) ReadCountFeedbacksByUserId(userId int64) (int, error) {
	var total int

	query := `
		SELECT COUNT(*)
		FROM feedback_model
		WHERE owner_id = $1
	`

	if err := service.DB.QueryRow(query, userId).Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}

func (s *FeedbackDbService) ReadAllByOwnerId(ownerId int64) ([]Feedback, error) {
	query := `
	SELECT id, title, description, slug, created_at, expires_at, board_id, owner_id, email, scope, resolution, updated_at, deleted_at
	FROM feedback_model
	WHERE owner_id = $1
	`

	rows, err := s.DB.Queryx(query, ownerId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var feedbacks []Feedback
	for rows.Next() {
		var feedback Feedback
		if err := rows.StructScan(&feedback); err != nil {
			return nil, err
		}

		feedbacks = append(feedbacks, feedback)
	}

	return feedbacks, nil
}

func (s *FeedbackDbService) ReadAllByBoardIdWithTx(tx *sqlx.Tx, id int64) ([]Feedback, error) {
	query := `
	SELECT id, title, description, slug, created_at, expires_at, board_id, owner_id, email, scope, resolution, updated_at, deleted_at
	FROM feedback_model
	WHERE board_id = $1
	`

	rows, err := tx.Queryx(query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var feedbacks []Feedback
	for rows.Next() {
		var feedback Feedback
		if err := rows.StructScan(&feedback); err != nil {
			return nil, err
		}

		feedbacks = append(feedbacks, feedback)
	}

	return feedbacks, nil
}

// reads all feedbacks that belongs to the user (including expired and deleted)
func (service *FeedbackDbService) ReadPaginatedAllFeedbacksByUserId(userId int64, page, perPage int) ([]Feedback, error) {
	offset := (page - 1) * perPage
	if offset < 1 {
		offset = 0 // safeguard
	}

	query := `
	SELECT id, title, description, slug, created_at, expires_at, board_id, owner_id, email, scope, resolution, updated_at, deleted_at
	FROM feedback_model
	WHERE owner_id = $1
	ORDER BY created_at DESC
	LIMIT $2 OFFSET $3
	`

	rows, err := service.DB.Query(query, userId, perPage, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var feedbacks []Feedback

	for rows.Next() {
		var f Feedback
		err := rows.Scan(
			&f.ID,
			&f.Title,
			&f.Description,
			&f.Slug,
			&f.CreatedAt,
			&f.ExpiresAt,
			&f.BoardID,
			&f.OwnerID,
			&f.Email,
			&f.Scope,
			&f.Resolution,
			&f.UpdatedAt,
			&f.DeletedAt,
		)

		if err != nil {
			return nil, err
		}

		feedbacks = append(feedbacks, f)
	}

	return feedbacks, nil

}

// ReadActivePaginatedFilteredFeedbacksByBoardId retrieves a paginated list of active (non-expired, non-deleted)
// feedback entries for a given board ID, optionally filtered by scope and resolution.
//
// Pagination is controlled via page and perPage, where page is 1-based. The offset is safely clamped to zero
// if an invalid page value results in a negative offset.
//
// Filtering behavior:
//   - If scope is not one of constants.FeedbackScopes, it is ignored and no scope filtering is applied.
//   - If resolution is not one of constants.FeedbackResolutions, it is ignored and no resolution filtering is applied.
//   - If either filter is an empty string, that filter is not applied in the SQL query.
//
// The function returns feedback entries ordered by expiration date (descending), limited to active records only:
//   - expires_at must be NULL
//   - deleted_at must be NULL
//
// Parameters:
//   - boardId: ID of the board to retrieve feedback for
//   - page: page number (1-based)
//   - perPage: number of records per page
//   - scope: optional scope filter (validated against allowed scopes)
//   - resolution: optional resolution filter (validated against allowed resolutions)
//
// Returns:
//   - []Feedback: slice of feedback records matching the query
//   - error: database or scan error if query execution fails
func (service *FeedbackDbService) ReadActivePaginatedFilteredFeedbacksByBoardId(boardId int64, page, perPage int, scope, resolution string) ([]Feedback, error) {
	offset := (page - 1) * perPage
	if offset < 1 {
		offset = 0 // safeguard
	}

	scopeFilter := ""
	if _, exist := constants.FeedbackScopes[scope]; exist {
		scopeFilter = scope
	}
	// if slices.Contains(constants.FeedbackScopes, scope) {
	// 	scopeFilter = scope
	// }

	resFilter := ""
	if _, exist := constants.FeedbackResolutions[resolution]; exist {
		resFilter = resolution
	}
	// if slices.Contains(constants.FeedbackResolutions, resolution) {
	// 	resFilter = resolution
	// }

	// WHERE ($1 = '' OR scope = $1)
	// returns everything otherwise return using the filter
	//
	query := `
	SELECT id, title, description, slug, created_at, expires_at, board_id, owner_id, email, scope, resolution, updated_at, deleted_at
	FROM feedback_model
	WHERE board_id = $1 
	AND expires_at is NULL 
	AND deleted_at IS NULL
	AND ($2 = '' OR scope = $2)
	AND ($3 = '' OR resolution = $3)
	ORDER BY created_at DESC
	LIMIT $4 OFFSET $5
	`

	rows, err := service.DB.Queryx(query, boardId, scopeFilter, resFilter, perPage, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var feedbacks []Feedback

	for rows.Next() {
		var f Feedback
		if err := rows.StructScan(&f); err != nil {
			return nil, err
		}

		// err := rows.Scan(
		// 	&f.ID,
		// 	&f.Title,
		// 	&f.Description,
		// 	&f.Slug,
		// 	&f.CreatedAt,
		// 	&f.ExpiresAt,
		// 	&f.BoardID,
		// 	&f.OwnerID,
		// 	&f.Email,
		// 	&f.Scope,
		// 	&f.Resolution,
		// 	&f.UpdatedAt,
		// 	&f.DeletedAt,
		// )

		// if err != nil {
		// 	return nil, err
		// }

		feedbacks = append(feedbacks, f)
	}
	// Check for errors that occurred during iteration
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return feedbacks, nil
}

// ReadActiveTotalCountFilteredFeedbacksByBoardId returns the total count of active feedback entries
// for a specified board ID, optionally filtered by scope and resolution.
//
// Only non-expired and non-deleted feedback is counted:
//   - expires_at must be NULL (active feedback)
//   - deleted_at must be NULL
//
// Filtering behavior:
//   - scope is validated against constants.FeedbackScopes. If invalid, it is treated as an empty filter.
//   - resolution is validated against constants.FeedbackResolutions. If invalid, it is treated as an empty filter.
//   - Empty filters are ignored in the SQL query via:
//     ($2 = ” OR scope = $2)
//     ($3 = ” OR resolution = $3)
//
// Parameters:
//   - boardId: ID of the board to count feedback for
//   - scope: optional scope filter (validated against allowed scopes)
//   - resolution: optional resolution filter (validated against allowed resolutions)
//
// Returns:
//   - int: total count of matching active feedback entries
//   - error: database error if query execution fails
func (service *FeedbackDbService) ReadActiveTotalCountFilteredFeedbacksByBoardId(boardId int64, scope, resolution string) (int, error) {
	var total int

	scopeFilter := ""
	if _, exist := constants.FeedbackScopes[scope]; exist {
		scopeFilter = scope
	}
	// if slices.Contains(constants.FeedbackScopes, scope) {
	// 	scopeFilter = scope
	// }

	resFilter := ""
	if _, exist := constants.FeedbackResolutions[resolution]; exist {
		resFilter = resolution
	}
	// if slices.Contains(constants.FeedbackResolutions, resolution) {
	// 	resFilter = resolution
	// }

	query := `
		SELECT COUNT(*)
		FROM feedback_model
		WHERE board_id = $1 
		AND expires_at is NULL 
		AND deleted_at IS NULL
		AND ($2 = '' OR scope = $2)
		AND ($3 = '' OR resolution = $3)
	`

	if err := service.DB.QueryRow(query, boardId, scopeFilter, resFilter).Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}

// ReadPaginatedPendingFeedbacksByBoardId retrieves a paginated list of pending (not expired) feedback entries
// for a given board ID.
//
// A feedback entry is considered pending if:
//   - deleted_at IS NULL (not deleted)
//   - expires_at IS NOT NULL (has an expiration time)
//   - expires_at > NOW() (has not expired yet)
//
// Results are ordered by expires_at in descending order (closest-to-expiry first)
// and returned using LIMIT/OFFSET pagination.
//
// Pagination is 1-based. The offset is calculated from the page number, and
// is clamped to zero if it becomes negative.
//
// Note: perPage controls the number of rows returned per page.
//
// Parameters:
//   - boardId: ID of the board to fetch feedback for
//   - page: 1-based page index
//   - perPage: number of records per page
//
// Returns:
//   - []Feedback: list of pending feedback entries
//   - error: database or row scanning error if query execution fails
func (service *FeedbackDbService) ReadPaginatedPendingFeedbacksByBoardId(boardId int64, page int, perPage int) ([]Feedback, error) {
	offset := (page - 1) * 10
	if offset < 1 {
		offset = 0
	}

	query := `
		SELECT id, title, description, slug, created_at, expires_at, board_id, owner_id, email, scope, resolution, updated_at, deleted_at
		FROM feedback_model
		WHERE board_id = $1
		AND deleted_at IS NULL
		AND expires_at IS NOT NULL
		AND expires_at > NOW()
		ORDER BY expires_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := service.DB.Query(query, boardId, perPage, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var feedbacks []Feedback

	for rows.Next() {
		var f Feedback
		err := rows.Scan(
			&f.ID,
			&f.Title,
			&f.Description,
			&f.Slug,
			&f.CreatedAt,
			&f.ExpiresAt,
			&f.BoardID,
			&f.OwnerID,
			&f.Email,
			&f.Scope,
			&f.Resolution,
			&f.UpdatedAt,
			&f.DeletedAt,
		)

		if err != nil {
			return nil, err
		}

		feedbacks = append(feedbacks, f)
	}

	return feedbacks, nil
}

// ReadPendingTotalCountFeedbacksByBoardId returns the total number of pending feedback entries
// for a given board ID.
//
// A feedback entry is considered pending if:
//   - deleted_at IS NULL (not deleted)
//   - expires_at IS NOT NULL (has an expiration timestamp)
//   - expires_at > NOW() (has not expired yet)
//
// This function performs a COUNT(*) query and returns the total number of
// matching records for the specified board.
//
// Parameters:
//   - boardId: ID of the board to count pending feedback for
//
// Returns:
//   - int: total number of pending feedback entries
//   - error: database error if the query execution fails
func (service *FeedbackDbService) ReadPendingTotalCountFeedbacksByBoardId(boardId int64) (int, error) {
	var total int
	err := service.DB.QueryRow(`
		SELECT COUNT(*)
		FROM feedback_model
		WHERE deleted_at IS NULL
		AND expires_at IS NOT NULL
		AND expires_at > NOW()
		AND board_id = $1
	`, boardId).Scan(&total)

	if err != nil {
		return 0, err
	}
	return total, nil
}

// ReadActiveFeedbackBySlug retrieves an active feedback entry by its unique slug.
//
// A feedback entry is considered active when:
//   - deleted_at IS NULL (not deleted)
//   - expires_at IS NULL (does not expire), or expires_at is in the future
//
// If no matching active feedback exists, or if the feedback has been deleted
// or expired, the underlying database query will return an error.
//
// Parameters:
//   - slug: unique feedback slug
//
// Returns:
//   - *Feedback: the matching active feedback record
//   - error: database error or sql.ErrNoRows if no active feedback is found
func (service *FeedbackDbService) ReadActiveFeedbackBySlug(slug string) (*Feedback, error) {

	query := `
	SELECT id, title, description, slug, created_at, expires_at, board_id, owner_id, email, scope, resolution, updated_at, deleted_at
	FROM feedback_model
	WHERE slug = $1
	AND deleted_at IS NULL
	AND (
		expires_at IS NULL
		OR expires_at > NOW()
	)
	`

	var feedback Feedback

	if err := service.DB.Get(&feedback, query, slug); err != nil {
		return nil, err
	}

	return &feedback, nil
}

func (s *FeedbackDbService) ReadBySlug(slug string) (*Feedback, error) {
	query := `
	SELECT id, title, description, slug, created_at, expires_at, board_id, owner_id, email, scope, resolution, updated_at, deleted_at
	FROM feedback_model
	WHERE slug = $1
	`

	var f Feedback
	if err := s.DB.Get(&f, query, slug); err != nil {
		return nil, err
	}

	return &f, nil
}

// ReadTotalActiveFeedbackCountByBoardId returns the total number of active feedback entries
// for a given board ID.
//
// A feedback entry is considered active if:
//   - deleted_at IS NULL (not deleted)
//   - expires_at IS NULL (no expiration date set)
//
// This function performs a COUNT(*) query and returns the number of matching
// active feedback records for the specified board.
//
// Parameters:
//   - boardId: ID of the board to count active feedback for
//
// Returns:
//   - int: total number of active feedback entries
//   - error: database error if the query execution fails
func (service *FeedbackDbService) ReadTotalActiveFeedbackCountByBoardId(boardId int64) (int, error) {
	var total int

	query := `
	SELECT COUNT(*)
	FROM feedback_model
	WHERE expires_at IS NULL AND deleted_at IS NULL
	AND board_id = $1
	`

	if err := service.DB.QueryRow(query, boardId).Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}

// ReadTotalScopeCountByBoardIdAndScopeValue returns the total number of active feedback entries
// for a given board ID that match a specific scope.
//
// A feedback entry is considered active if:
//   - deleted_at IS NULL (not deleted)
//   - expires_at IS NULL (no expiration date)
//
// Only feedback entries with scope equal to the provided value are counted.
//
// Parameters:
//   - boardId: ID of the board to filter feedback by
//   - scope: the scope value to filter feedback entries
//
// Returns:
//   - int: total number of active feedback entries with the given scope
//   - error: database error if the query execution fails
func (service *FeedbackDbService) ReadTotalScopeCountByBoardIdAndScopeValue(boardId int64, scope string) (int, error) {
	var total int
	err := service.DB.QueryRow(`
		SELECT COUNT(*)
		FROM feedback_model
		WHERE expires_at IS NULL
		AND deleted_at IS NULL
		AND scope=$1
		AND board_id = $2
	`, scope, boardId).Scan(&total)

	if err != nil {
		return 0, err
	}
	return total, nil
}

// ReadTotalResolutionCountByBoardIdAndResValue returns the total number of active feedback entries
// for a given board ID that match a specific resolution value.
//
// A feedback entry is considered active if:
//   - deleted_at IS NULL (not deleted)
//   - expires_at IS NULL (no expiration date)
//
// Only feedback entries with resolution equal to the provided value are counted.
//
// Parameters:
//   - boardId: ID of the board to filter feedback by
//   - resolution: the resolution value to filter feedback entries
//
// Returns:
//   - int: total number of active feedback entries with the given resolution
//   - error: database error if the query execution fails
func (service *FeedbackDbService) ReadTotalResolutionCountByBoardIdAndResValue(boardId int64, resolution string) (int, error) {
	var total int
	err := service.DB.QueryRow(`
		SELECT COUNT(*)
		FROM feedback_model
		WHERE expires_at IS NULL
		AND deleted_at IS NULL
		AND resolution=$1
		AND board_id = $2
	`, resolution, boardId).Scan(&total)

	if err != nil {
		return 0, err
	}
	return total, nil
}

// IsTherePending checks whether there are any pending feedback entries for a given board ID.
//
// A feedback entry is considered pending if:
//   - deleted_at IS NULL (not deleted)
//   - expires_at IS NOT NULL (has an expiration timestamp)
//   - expires_at > NOW() (has not expired yet)
//
// This function performs a COUNT(*) query and returns true if at least one
// matching pending feedback exists.
//
// Parameters:
//   - boardId: ID of the board to check for pending feedback
//
// Returns:
//   - bool: true if there is at least one pending feedback, false otherwise
//   - error: database error if the query execution fails
func (service *FeedbackDbService) IsTherePending(boardId int64) (bool, error) {
	var total int
	err := service.DB.QueryRow(`
		SELECT COUNT(*)
		FROM feedback_model
		WHERE expires_at IS NOT NULL
		AND expires_at > NOW()
		AND deleted_at IS NULL
		AND board_id = $1
	`, boardId).Scan(&total)

	if err != nil {
		return false, err
	}

	if total > 0 {
		return true, nil
	}
	return false, nil
}

// UpdateFeedback updates the scope and resolution of an existing feedback entry.
//
// During the update:
//   - scope and resolution are set to the provided values (converted to lowercase)
//   - deleted_at is cleared (set to NULL)
//   - updated_at is set to the current UTC time
//   - expires_at is cleared (set to NULL)
//
// The function performs an UPDATE query on the feedback_model table and returns
// the updated feedback record.
//
// Parameters:
//   - f: pointer to the Feedback struct to update (must have a valid ID)
//   - scope: new scope value to set
//   - resolution: new resolution value to set
//
// Returns:
//   - *Feedback: the updated feedback record after applying changes
//   - error: database error if the query execution fails
func (service *FeedbackDbService) UpdateFeedback(f *Feedback, scope, resolution string) error {
	f.Scope = sql.NullString{
		String: strings.ToLower(scope),
		Valid:  true,
	}

	f.Resolution = sql.NullString{
		String: strings.ToLower(resolution),
		Valid:  true,
	}

	f.DeletedAt = sql.NullTime{
		Valid: false,
	}

	f.UpdatedAt = sql.NullTime{
		Time:  time.Now().UTC(),
		Valid: true,
	}

	f.ExpiresAt = sql.NullTime{
		Valid: false,
	}

	query := `
	UPDATE feedback_model
	SET scope = $1, resolution = $2, deleted_at = $3, updated_at = $4, expires_at = $5
	WHERE id = $6
	RETURNING id, title, description, slug, expires_at, board_id, owner_id, email, scope, resolution, updated_at, deleted_at
	`
	if err := service.DB.Get(
		f,
		query,
		f.Scope,
		f.Resolution,
		f.DeletedAt,
		f.UpdatedAt,
		f.ExpiresAt,
		f.ID,
	); err != nil {
		return err
	}

	return nil

}

func (s *FeedbackDbService) UpdateScopeWihTx(tx *sqlx.Tx, f *Feedback, scope string) error {
	scopeNullString := sql.NullString{
		String: strings.ToLower(scope),
		Valid:  true,
	}

	query := `
	UPDATE feedback_model
	SET scope = $1, updated_at = NOW()
	WHERE id = $2
	RETURNING id, title, description, slug, expires_at, board_id, owner_id, email, scope, resolution, updated_at, deleted_at
	`

	if err := tx.Get(
		f,
		query,
		scopeNullString,
		f.ID,
	); err != nil {
		return err
	}

	return nil
}

func (s *FeedbackDbService) UpdateResolutionWihTx(tx *sqlx.Tx, f *Feedback, resolution string) error {
	resNullString := sql.NullString{
		String: strings.ToLower(resolution),
		Valid:  true,
	}

	query := `
	UPDATE feedback_model
	SET resolution = $1, updated_at = NOW()
	WHERE id = $2
	RETURNING id, title, description, slug, expires_at, board_id, owner_id, email, scope, resolution, updated_at, deleted_at
	`

	if err := tx.Get(
		f,
		query,
		resNullString,
		f.ID,
	); err != nil {
		return err
	}

	return nil
}

func (s *FeedbackDbService) UpdateExpiryWihTx(tx *sqlx.Tx, f *Feedback, exp sql.NullTime) error {
	query := `
	UPDATE feedback_model
	SET expires_at = $1, updated_at = NOW()
	WHERE id = $2
	RETURNING id, title, description, slug, expires_at, board_id, owner_id, email, scope, resolution, updated_at, deleted_at
	`

	if err := tx.Get(
		f,
		query,
		exp,
		f.ID,
	); err != nil {
		return err
	}

	return nil
}

func (s *FeedbackDbService) UpdateOwnerIdWithTx(tx *sqlx.Tx, feedback *Feedback, id int64) error {
	query := `
	UPDATE feedback_model
	SET owner_id = NULLIF($1, 0)
	WHERE id = $2
	RETURNING id, title, description, slug, expires_at, board_id, owner_id, email, scope, resolution, updated_at, deleted_at
	`

	if err := tx.Get(feedback, query, id, feedback.ID); err != nil {
		return err
	}

	return nil
}

func (s *FeedbackDbService) UpdateEmailWithTx(tx *sqlx.Tx, feedback *Feedback, email string) error {
	query := `
	UPDATE feedback_model
	SET email = NULLIF($1, '')
	WHERE id = $2
	RETURNING id, title, description, slug, expires_at, board_id, owner_id, email, scope, resolution, updated_at, deleted_at
	`

	if err := tx.Get(feedback, query, email, feedback.ID); err != nil {
		return err
	}

	return nil
}

// DeleteFeedback performs a soft delete on a feedback entry by setting its deleted_at timestamp.
//
// Instead of permanently removing the record, this function marks the feedback as deleted by:
//   - setting deleted_at to the current UTC time
//
// The function then updates the database and returns the updated feedback record.
//
// Parameters:
//   - f: pointer to the Feedback struct to delete (must have a valid ID)
//
// Returns:
//   - *Feedback: the updated feedback record with deleted_at set
//   - error: database error if the update query fails
func (service *FeedbackDbService) DeleteFeedback(f *Feedback) error {
	f.DeletedAt = sql.NullTime{
		Time:  time.Now().UTC(),
		Valid: true,
	}

	query := `
	UPDATE feedback_model
	SET deleted_at = $1
	WHERE id = $2
	RETURNING id, title, description, slug, expires_at, board_id, owner_id, email, scope, resolution, updated_at, deleted_at
	`
	err := service.DB.Get(
		f,
		query,
		f.DeletedAt,
		f.ID,
	)

	if err != nil {
		return err
	}

	return nil
}

func (service *FeedbackDbService) DeleteFeedbackWithTx(tx *sqlx.Tx, f *Feedback) error {
	f.DeletedAt = sql.NullTime{
		Time:  time.Now().UTC(),
		Valid: true,
	}

	query := `
	UPDATE feedback_model
	SET deleted_at = $1
	WHERE id = $2
	RETURNING id, title, description, slug, expires_at, board_id, owner_id, email, scope, resolution, updated_at, deleted_at
	`
	if err := tx.Get(
		f,
		query,
		f.DeletedAt,
		f.ID,
	); err != nil {
		return err
	}

	return nil
}
