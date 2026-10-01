package sessions_utils

import (
	"app/internal/constants"
	authid_models "app/internal/models/authid"
	user_models "app/internal/models/user"
	"errors"
	"net/http"

	"github.com/alexedwards/scs/v2"
	"github.com/jmoiron/sqlx"
)

type SessionService struct {
	Request *http.Request
	Session *scs.SessionManager
}

var ErrNoUserInSession = errors.New("no user in session")

func NewSessionService(r *http.Request, s *scs.SessionManager) *SessionService {
	return &SessionService{
		Request: r,
		Session: s,
	}
}

// Grabs user session
// check that user exist
// check that user is not deactivated
func (s *SessionService) GetUserFromSession(db *sqlx.DB) (*user_models.User, error) {
	userDbService := user_models.NewUserDbService(db)
	authidDbService := authid_models.NewAuthidDbService(db)

	authidValue, ok := s.Session.Get(s.Request.Context(), constants.SessionManagerSessionKey).(string)
	if !ok {
		return nil, ErrNoUserInSession
	}

	authid, err := authidDbService.ReadByValue(authidValue)
	if err != nil {
		if errDestroy := s.DestroySession(); errDestroy != nil {
			return nil, errDestroy
		}
		return nil, err
	}

	user, err := userDbService.ReadByAuthidId(authid.ID)
	if err != nil {
		if errDestroy := s.DestroySession(); errDestroy != nil {
			return nil, errDestroy
		}
		return nil, err
	}

	if user.DeactivatedAt.Valid {
		if errDestroy := s.DestroySession(); errDestroy != nil {
			return nil, errDestroy
		}
		return nil, ErrNoUserInSession
	}

	return user, nil

}

func (s *SessionService) ToFlash(msg string) {
	s.Session.Put(s.Request.Context(), constants.FlashSessionKey, msg)
}

func (s *SessionService) Login(authidValue string, remember bool) {
	//RenewToken prevents "Session Fixation" attacks by generating a totally brand new ID card.
	s.Session.RenewToken(s.Request.Context())
	s.Session.Put(s.Request.Context(), constants.SessionManagerSessionKey, authidValue)
	s.Session.RememberMe(s.Request.Context(), remember)
}

func (s *SessionService) DestroySession() error {
	// set empty token in the client
	err := s.Session.RenewToken(s.Request.Context())
	if err != nil {
		return err
	}

	// remove from backend
	s.Session.Remove(s.Request.Context(), constants.SessionManagerSessionKey)
	return nil
}
