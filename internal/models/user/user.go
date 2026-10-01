package user_models

import (
	"app/internal/constants"
	authid_models "app/internal/models/authid"
	randoms_utils "app/internal/utilities/randoms"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

// table name is user_model
type User struct {
	ID            int64        `db:"id"`
	Firstname     string       `db:"firstname"`
	Lastname      string       `db:"lastname"`
	Email         string       `db:"email"`
	Password      string       `db:"password"`
	Pid           string       `db:"pid"`
	AuthidId      int64        `db:"authid_id"`
	VerifiedAt    sql.NullTime `db:"verified_at"`
	DeactivatedAt sql.NullTime `db:"deactivated_at"`
	CreatedAt     time.Time    `db:"created_at"`
	UpdatedAt     sql.NullTime `db:"updated_at"`
}

type UserDbService struct {
	DB *sqlx.DB
}

func NewUserDbService(db *sqlx.DB) *UserDbService {
	return &UserDbService{
		DB: db,
	}
}

func (s *UserDbService) CreateWithTx(tx *sqlx.Tx, firstname, lastname, email, password string, authidId int64) (*User, error) {

	var pid string
	for range 5 {
		newPid, err := randoms_utils.GeneratePidValue()
		if err != nil {
			return nil, err
		}

		var exists bool
		query := `
		SELECT EXISTS (
			SELECT 1
			FROM user_model
			WHERE pid = $1
		);
		`

		err = tx.QueryRow(query, newPid).Scan(&exists)
		if err != nil {
			return nil, err
		}

		if exists {
			continue
		} else {
			pid = newPid
			break
		}
	}
	if pid == "" {
		return nil, errors.New("failed to generate pid")
	}

	// lower and trim
	email = strings.ToLower(strings.TrimSpace(email))

	hashedByte, err := bcrypt.GenerateFromPassword([]byte(password), constants.BcryptCost)
	if err != nil {
		return nil, err
	}

	query := `
	INSERT INTO user_model (firstname, lastname, email, password, pid, authid_id)
	VALUES ($1, $2, $3, $4, $5, $6)
	RETURNING id, firstname, lastname, email, password, pid, authid_id, verified_at, deactivated_at, created_at, updated_at
	`
	var user User
	if err := tx.Get(&user, query, firstname, lastname, email, string(hashedByte), pid, authidId); err != nil {
		return nil, err
	}

	return &user, nil
}

// func (service *UserDbService) Create(firstname, lastname, email, password string) (*User, error) {
// 	authidDbService := authid_models.NewAuthidDbService(service.DB)

// 	tx, err := service.DB.Beginx()
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer tx.Rollback() // happens when function exists

// 	// create authid - its a different model hence the use of tx
// 	authidId := int64(0)
// 	for range 5 {
// 		authidValue, err := randoms_utils.SecureRandString(32, 64)
// 		if err != nil {
// 			return nil, err
// 		}

// 		authid, err := authidDbService.CreateWithTx(tx, authidValue)
// 		if err != nil {
// 			continue
// 		}

// 		authidId = authid.ID
// 		break
// 	}
// 	if authidId == int64(0) {
// 		return nil, errors.New("failed to generate authid")
// 	}

// 	// different from authid
// 	// the value is part of the user table
// 	var pid string
// 	for range 5 {
// 		newPid, err := randoms_utils.SecureRandString(16, 32)
// 		if err != nil {
// 			return nil, err
// 		}

// 		var exists bool
// 		query := `
// 		SELECT EXISTS (
// 			SELECT 1
// 			FROM user_model
// 			WHERE pid = $1
// 		);
// 		`

// 		err = tx.QueryRow(query, strings.ToLower(newPid)).Scan(&exists)
// 		if err != nil {
// 			return nil, err
// 		}

// 		if exists {
// 			continue
// 		} else {
// 			pid = strings.ToLower(newPid)
// 			break
// 		}
// 	}
// 	if pid == "" {
// 		return nil, errors.New("failed to generate pid")
// 	}

// 	// lower and trim
// 	email = strings.ToLower(strings.TrimSpace(email))

// 	hashedByte, err := bcrypt.GenerateFromPassword([]byte(password), 12)
// 	if err != nil {
// 		return nil, err
// 	}

// 	query := `
// 	INSERT INTO user_model (firstname, lastname, email, password, pid, authid_id)
// 	VALUES ($1, $2, $3, $4, $5, $6)
// 	RETURNING id, firstname, lastname, email, password, pid, authid_id, verified_at, deactivated_at, created_at, updated_at
// 	`
// 	var user User
// 	if err := tx.Get(&user, query, firstname, lastname, email, string(hashedByte), pid, authidId); err != nil {
// 		return nil, err
// 	}

// 	if err := tx.Commit(); err != nil {
// 		return nil, err
// 	}

// 	//commit success, it will rollback but since its already committed, nothing to rollback
// 	return &user, nil
// }

func (s *UserDbService) ReadByAuthidId(authidId int64) (*User, error) {
	query := `
	SELECT id, firstname, lastname, email, password, pid, authid_id, verified_at, deactivated_at, updated_at
	FROM user_model
	WHERE authid_id=$1
	`

	var user User
	if err := s.DB.Get(&user, query, authidId); err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *UserDbService) ReadById(id int64) (*User, error) {
	query := `
	SELECT id, firstname, lastname, email, password, pid, authid_id, verified_at, deactivated_at, updated_at
	FROM user_model
	WHERE id=$1
	`

	var user User
	if err := s.DB.Get(&user, query, id); err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *UserDbService) ReadByEmail(email string) (*User, error) {
	query := `
	SELECT id, firstname, lastname, email, password, pid, authid_id, verified_at, deactivated_at, updated_at
	FROM user_model
	WHERE email=$1
	`

	var user User
	if err := s.DB.Get(&user, query, email); err != nil {
		return nil, err
	}
	return &user, nil
}

func (service *UserDbService) ReadActiveUserById(id int64) (*User, error) {
	query := `
	SELECT id, firstname, lastname, email, password, pid, authid_id, verified_at, deactivated_at, updated_at
	FROM user_model
	WHERE id=$1
	AND deactivated_at IS NULL
	`
	var user User
	err := service.DB.Get(&user, query, id)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (service *UserDbService) ReadActiveUserByAuthidId(authidId int64) (*User, error) {
	query := `
	SELECT id, firstname, lastname, email, password, pid, authid_id, verified_at, deactivated_at, updated_at
	FROM user_model
	WHERE authid_id=$1
	`
	var user User
	if err := service.DB.Get(&user, query, authidId); err != nil {
		return nil, err
	}

	return &user, nil
}

func (service *UserDbService) ReadUserByEmail(email string) (*User, error) {
	query := `
	SELECT id, firstname, lastname, email, password, pid, authid_id, verified_at, deactivated_at, updated_at
	FROM user_model
	WHERE email=$1
	`
	var user User
	err := service.DB.Get(&user, query, email)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (service *UserDbService) IsEmailAvailable(email string) (bool, error) {
	var exists bool
	query := `
	SELECT EXISTS (
		SELECT 1
		FROM user_model
		WHERE email = $1
	);
	`
	err := service.DB.QueryRow(query, email).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (service *UserDbService) UpdateName(user *User, firstname, lastname string) error {
	query := `
	UPDATE user_model
	SET firstname = $1, lastname = $2, updated_at = NOW()
	WHERE id = $3
	RETURNING id, firstname, lastname, email, password, pid, authid_id, verified_at, deactivated_at, created_at, updated_at
	`
	if err := service.DB.Get(user, query, firstname, lastname, user.ID); err != nil {
		return err
	}

	return nil
}

func (s *UserDbService) UpdateFirstnameWithTx(tx *sqlx.Tx, user *User, firstname string) error {
	query := `
	UPDATE user_model
	SET firstname = $1, updated_at = NOW()
	WHERE id = $2
	RETURNING id, firstname, lastname, email, password, pid, authid_id, verified_at, deactivated_at, created_at, updated_at
	`

	if err := tx.Get(user, query, firstname, user.ID); err != nil {
		return err
	}

	return nil
}

func (s *UserDbService) UpdateLastnameWithTx(tx *sqlx.Tx, user *User, lastname string) error {
	query := `
	UPDATE user_model
	SET lastname = $1, updated_at = NOW()
	WHERE id = $2
	RETURNING id, firstname, lastname, email, password, pid, authid_id, verified_at, deactivated_at, created_at, updated_at
	`

	if err := tx.Get(user, query, lastname, user.ID); err != nil {
		return err
	}

	return nil
}

func (service *UserDbService) UpdateEmail(user *User, email string) error {
	query := `
	UPDATE user_model
	SET email = $1, verified_at = NULL, updated_at = NOW()
	WHERE id = $2
	RETURNING id, firstname, lastname, email, password, pid, authid_id, verified_at, deactivated_at, created_at, updated_at
	`
	if err := service.DB.Get(user, query, email, user.ID); err != nil {
		return err
	}

	return nil
}

func (s *UserDbService) UpdateEmailWithTx(tx *sqlx.Tx, user *User, email string) error {
	query := `
	UPDATE user_model
	SET email = $1, updated_at = NOW()
	WHERE id = $2
	RETURNING id, firstname, lastname, email, password, pid, authid_id, verified_at, deactivated_at, created_at, updated_at
	`

	if err := tx.Get(user, query, email, user.ID); err != nil {
		return err
	}

	return nil
}

func (s *UserDbService) UpdateVerifiedWithTx(tx *sqlx.Tx, user *User, verified *sql.NullTime) error {
	query := `
	UPDATE user_model
	SET verified_at = $1, updated_at = NOW()
	WHERE id = $2
	RETURNING id, firstname, lastname, email, password, pid, authid_id, verified_at, deactivated_at, created_at, updated_at
	`

	if err := tx.Get(user, query, verified, user.ID); err != nil {
		return err
	}

	return nil
}

func (s *UserDbService) UpdateVerified(user *User, verified *sql.NullTime) error {
	query := `
	UPDATE user_model
	SET verified_at = $1, updated_at = NOW()
	WHERE id = $2
	RETURNING id, firstname, lastname, email, password, pid, authid_id, verified_at, deactivated_at, created_at, updated_at
	`

	if err := s.DB.Get(user, query, verified, user.ID); err != nil {
		return err
	}

	return nil
}

func (service *UserDbService) UpdatePasswordWithTx(tx *sqlx.Tx, user *User, password string) error {
	// authidDbService := authid_models.NewAuthidDbService(service.DB)
	// tx, err := service.DB.Beginx()
	// if err != nil {
	// 	return nil, err
	// }
	// defer tx.Rollback()

	// // create new authid (used in session)
	// // change it for user global logout
	// authidId := int64(0)
	// for range 5 {
	// 	authidValue, err := randoms_utils.SecureRandString(32, 64)
	// 	if err != nil {
	// 		return nil, err
	// 	}

	// 	authid, err := authidDbService.CreateWithTx(tx, authidValue)
	// 	if err != nil {
	// 		continue
	// 	}

	// 	authidId = authid.ID
	// 	break
	// }
	// if authidId == int64(0) {
	// 	return nil, errors.New("failed to generate authid")
	// }

	hashedByte, err := bcrypt.GenerateFromPassword([]byte(password), constants.BcryptCost)
	if err != nil {
		return err
	}

	query := `
	UPDATE user_model
	SET password = $1, updated_at = NOW()
	WHERE id = $2
	RETURNING id, firstname, lastname, email, password, pid, authid_id, verified_at, deactivated_at, created_at, updated_at
	`

	if err := tx.Get(user, query, string(hashedByte), user.ID); err != nil {
		return err
	}

	//commit success, it will rollback but since its already committed, nothing to rollback
	return nil

}

func (service *UserDbService) UpdateReactivate(user *User) error {
	query := `
	UPDATE user_model
	SET deactivated_at = NULL
	WHERE id = $1
	RETURNING id, firstname, lastname, email, password, pid, authid_id, verified_at, deactivated_at, created_at, updated_at
	`

	if err := service.DB.Get(user, query, user.ID); err != nil {
		return err
	}

	return nil

}

func (service *UserDbService) UpdateDeactivate(user *User) error {
	query := `
	UPDATE user_model
	SET deactivated_at = NOW()
	WHERE id = $1
	RETURNING id, firstname, lastname, email, password, pid, authid_id, verified_at, deactivated_at, created_at, updated_at
	`
	var u User
	if err := service.DB.Get(&u, query, user.ID); err != nil {
		return err
	}

	return nil

}

func (service *UserDbService) UpdateAuthidIdWithTx(tx *sqlx.Tx, user *User, authid *authid_models.Authid) error {
	query := `
	UPDATE user_model
	SET authid_id = $1
	WHERE id = $2
	RETURNING id, firstname, lastname, email, password, pid, authid_id, verified_at, deactivated_at, created_at, updated_at
	`

	var u User
	if err := tx.Get(&u, query, authid.ID, user.ID); err != nil {
		return err
	}

	return nil
}

func (service *UserDbService) DeleteUserWithTx(tx *sqlx.Tx, user *User) error {
	query := `
	DELETE FROM user_model
	WHERE id = $1
	`

	res, err := tx.Exec(query, user.ID)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("no user found with id %d", user.ID)
	}

	return nil
}
