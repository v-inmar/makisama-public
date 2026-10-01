package authid_models

import (
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
)

// authid_model table
type Authid struct {
	ID        int64        `db:"id"`
	Value     string       `db:"value"`
	TTL       sql.NullTime `db:"ttl"`
	CreatedAt time.Time    `db:"created_at"`
}

type AuthidDbService struct {
	DB *sqlx.DB
}

func NewAuthidDbService(db *sqlx.DB) *AuthidDbService {
	return &AuthidDbService{
		DB: db,
	}
}

// Create inserts a new Authid record into the database using the provided
// transaction and returns the created Authid with all populated fields.
//
// The function executes an INSERT statement with a RETURNING clause to fetch
// the generated id, value, ttl, and created_at fields from the database.
//
// The caller is responsible for managing the transaction lifecycle, including
// committing or rolling back the transaction.
//
// Returns the created Authid on success, or an error if the insert fails
// or if scanning the returned row fails.
func (service *AuthidDbService) CreateWithTx(tx *sqlx.Tx, value string) (*Authid, error) {
	query := `
	INSERT INTO authid_model (value)
	VALUES ($1)
	RETURNING id, value, ttl, created_at
	`

	var authid Authid
	if err := tx.Get(&authid, query, value); err != nil {
		return nil, err
	}

	return &authid, nil
}

func (service *AuthidDbService) ReadByIdWithTx(tx *sqlx.Tx, id int64) (*Authid, error) {
	query := `
	SELECT id, value, ttl, created_at
	FROM authid_model
	WHERE id = $1
	`
	var authid Authid
	err := tx.Get(&authid, query, id)
	if err != nil {
		return nil, err
	}

	return &authid, nil
}

// ReadById retrieves an Authid record from the database by its ID.
//
// It executes a SQL query against the authid_model table and scans the result
// into an Authid struct. If no record is found or the query fails, an error
// is returned.
//
// Parameters:
//   - id: The unique identifier of the Authid record to retrieve.
//
// Returns:
//   - *Authid: A pointer to the retrieved Authid struct if successful.
//   - error: An error if the query fails or no record is found.
func (service *AuthidDbService) ReadById(id int64) (*Authid, error) {
	query := `
	SELECT id, value, ttl, created_at
	FROM authid_model
	WHERE id = $1
	`
	var authid Authid
	err := service.DB.Get(&authid, query, id)
	if err != nil {
		return nil, err
	}

	return &authid, nil
}

// ReadByValue retrieves an Authid record from the database by its value.
//
// It queries the authid_model table for a record matching the provided value
// and scans the result into an Authid struct.
//
// Parameters:
//   - value: The unique AuthID value to search for.
//
// Returns:
//   - *Authid: A pointer to the matching Authid record if found.
//   - error: An error if the query fails or no matching record exists.
func (service *AuthidDbService) ReadByValue(value string) (*Authid, error) {
	query := `
	SELECT id, value, ttl, created_at
	FROM authid_model
	WHERE value = $1
	`
	var authid Authid
	if err := service.DB.Get(&authid, query, value); err != nil {
		return nil, err
	}

	return &authid, nil
}

func (service *AuthidDbService) UpdateTTLWithTx(tx *sqlx.Tx, authid *Authid, ttl *time.Time) error {
	query := `
	UPDATE authid_model
	SET ttl = $1
	WHERE id = $2
	RETURNING id, value, ttl, created_at
	`

	if err := tx.Get(authid, query, ttl, authid.ID); err != nil {
		return err
	}

	return nil
}
