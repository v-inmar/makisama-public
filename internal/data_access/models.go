package dataaccess

import (
	"database/sql"
	"time"
)

type ApiResponseDao struct {
	Code    int    `json:"code"`
	Status  string `json:"status"`
	Payload any    `json:"payload"`
}

type CreateFeedbackResponseDao struct {
	Message string `json:"message"`
}

type CreateFeedbackRequestDao struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Email       string `json:"email"`
	BoardSlug   string `json:"boardSlug"`
}

type ApiCreateFeedbackRequestDao struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	BoardSlug   string `json:"boardSlug"`
}

type ReadFeedbackResponseDao struct {
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	Slug         string    `json:"slug"`
	Resolution   string    `json:"resolution"`
	Scope        string    `json:"scope"`
	ExpiresAt    string    `json:"expires_at"`
	IsBoardOwner bool      `json:"is_board_owner"`
	CreatedAt    time.Time `json:"created_at"`
	BoardSlug    string    `json:"board_slug"`
}

type ReadVersionResponseDao struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildTime string `json:"build_time"`
}

// type CreateFeedbackDao struct {
// 	Title       string `json:"title"`
// 	Description string `json:"description"`
// 	Email       string `json:"email"`
// 	BoardSlug   string `json:"boardSlug"`
// }

// type BoardStatsDao struct {
// 	TotalNumberFeedback int `json:"total_number_feedback"`
// 	Features            int `json:"features"`
// 	Bugs                int `json:"bugs"`
// 	Improvements        int `json:"improvements"`
// 	UnderDiscussion     int `json:"under_discussion"`
// 	InProgress          int `json:"in_progress"`
// 	Released            int `json:"released"`
// 	OutOfScope          int `json:"out_of_scope"`
// 	Rejected            int `json:"rejected"`
// 	Duplicate           int `json:"duplicate"`
// }

type BoardFeedbackStrippedDao struct {
	Title       string
	Description string
	Slug        string
	CreatedAt   time.Time
	Scope       string
	Resolution  string
}

type PendingFeedbackStrippedDao struct {
	Title       string
	Description string
	Slug        string
	CreatedAt   time.Time
	ExpiresAt   time.Time
}

type BoardListDataDao struct {
	Title         string
	Description   string
	Slug          string
	CreatedAt     time.Time
	FeedbackCount int
}

type FeedbackListDataDao struct {
	BoardSlug          string
	BoardStatus        string
	FeedbackStatus     string
	FeedbackSlug       string
	FeedbackTitle      string
	FeedbackCreatedAt  time.Time
	FeedbackScope      sql.NullString
	FeedbackResolution sql.NullString
	FeedbackDeletedAt  sql.NullTime
	FeedbackExpiresAt  sql.NullTime
}
