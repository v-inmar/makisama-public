package handlers

import (
	"html/template"

	"github.com/alexedwards/scs/v2"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type HandlerRepo struct {
	Templates map[string]*template.Template
	Session   *scs.SessionManager
	DB        *sqlx.DB
	Logger    *zap.SugaredLogger
}

func NewRepo(t map[string]*template.Template, s *scs.SessionManager, db *sqlx.DB, l *zap.SugaredLogger) *HandlerRepo {
	return &HandlerRepo{
		Templates: t,
		Session:   s,
		DB:        db,
		Logger:    l,
	}
}
