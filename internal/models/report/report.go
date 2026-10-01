package report_models

import (
	"app/internal/constants"
	randoms_utils "app/internal/utilities/randoms"
	"database/sql"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
)

type ReportBoard struct {
	ID        int64          `db:"id"`
	ReportID  string         `db:"report_id"`
	Reason    string         `db:"reason"`
	Comment   sql.NullString `db:"comment"`
	Email     sql.NullString `db:"email"`
	BoardId   int64          `db:"board_id"`
	CreatedAt time.Time      `db:"created_at"`
}

type ReportBoardDbService struct {
	DB *sqlx.DB
}

func NewReportBoardDbService(db *sqlx.DB) *ReportBoardDbService {
	return &ReportBoardDbService{
		DB: db,
	}
}

func (service *ReportBoardDbService) Create(reason, comment, email string, boardId int64) (*ReportBoard, error) {
	var reportId string
	for range 5 {
		rId, err := randoms_utils.SecureRandStringWithPool(constants.AlphaNumericLowerCharset, 8, 8)
		if err != nil {
			continue
		}

		rIdQuery := `SELECT EXISTS ( SELECT 1 FROM report_board_model WHERE report_id = $1);`

		var exists bool
		if err := service.DB.QueryRow(rIdQuery, rId).Scan(&exists); err != nil {
			continue
		}

		if !exists {
			reportId = rId
			break
		}

	}

	if reportId == "" {
		return nil, errors.New("failed to generate report id")
	}

	var report ReportBoard

	insertQuery := `
	INSERT INTO report_board_model (report_id, reason, comment, email, board_id)
	VALUES ($1, $2, NULLIF($3, ''), NULLIF($4, ''), $5)
	RETURNING id, report_id, reason, comment, email, board_id, created_at
	`
	if err := service.DB.Get(&report, insertQuery, reportId, reason, comment, email, boardId); err != nil {
		return nil, err
	}

	return &report, nil

}

type ReportFeedback struct {
	ID         int64          `db:"id"`
	ReportID   string         `db:"report_id"`
	Reason     string         `db:"reason"`
	Comment    sql.NullString `db:"comment"`
	Email      sql.NullString `db:"email"`
	FeedbackId int64          `db:"feedback_id"`
	CreatedAt  time.Time      `db:"created_at"`
}

type ReportFeedbackDbService struct {
	DB *sqlx.DB
}

func NewReportFeedbackDbService(db *sqlx.DB) *ReportFeedbackDbService {
	return &ReportFeedbackDbService{
		DB: db,
	}
}

func (service *ReportFeedbackDbService) Create(reason, comment, email string, feedbackId int64) (*ReportFeedback, error) {
	var reportId string
	for range 5 {
		rId, err := randoms_utils.SecureRandStringWithPool(constants.AlphaNumericLowerCharset, 8, 8)
		if err != nil {
			continue
		}

		rIdQuery := `SELECT EXISTS ( SELECT 1 FROM report_feedback_model WHERE report_id = $1);`

		var exists bool
		if err := service.DB.QueryRow(rIdQuery, rId).Scan(&exists); err != nil {
			continue
		}

		if !exists {
			reportId = rId
			break
		}

	}

	if reportId == "" {
		return nil, errors.New("failed to generate report id")
	}

	var report ReportFeedback

	insertQuery := `
	INSERT INTO report_feedback_model (report_id, reason, comment, email, feedback_id)
	VALUES ($1, $2, NULLIF($3, ''), NULLIF($4, ''), $5)
	RETURNING id, report_id, reason, comment, email, feedback_id, created_at
	`
	if err := service.DB.Get(&report, insertQuery, reportId, reason, comment, email, feedbackId); err != nil {
		return nil, err
	}

	return &report, nil
}
