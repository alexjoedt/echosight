package http

import (
	"context"
	"net/http"
	"time"

	echosight "github.com/alexjoedt/echosight/internal"
	"github.com/alexjoedt/echosight/internal/logger"
	"github.com/alexjoedt/echosight/internal/validator"
)

func (s *Server) ServerContext(parent context.Context) (context.Context, context.CancelFunc) {
	if parent == nil {
		parent = context.Background()
	}

	if s.IsDev {
		return context.WithCancel(parent)
	}
	return context.WithTimeout(parent, time.Second*30)
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) error {

	var payload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Remember bool   `json:"remember"`
	}

	err := readJSON(r, &payload)
	if err != nil {
		s.log.Errorc("invalid payload", err)
		return echosight.ErrInvalidf("invalid payload")
	}

	v := validator.New()
	echosight.ValidateEmail(v, payload.Email)

	if !v.Valid() {
		writeJSON(w, http.StatusBadRequest, Response{
			Message: "invalid login",
			Errors:  v.Errors,
		})

		s.log.Errorw("invalid email", logger.Str("email", payload.Email))
		return echosight.ErrInvalidf("invalid email")
	}

	user, err := s.UserService.GetByEmail(r.Context(), payload.Email)
	if err != nil {
		s.log.Errorw("invalid login", logger.Str("error", err.Error()), logger.Str("email", payload.Email))
		return echosight.ErrUnauthorizedf("invalid login")
	}

	match, err := user.Password.Matches(payload.Password)
	if err != nil || !match {
		return echosight.ErrUnauthorizedf("invalid login")
	}

	session, err := s.newSession(user.ID)
	if err != nil {
		return echosight.ErrInternalf("failed to create session").WithError(err)
	}

	err = s.SessionService.Put(r.Context(), session)
	if err != nil {
		return echosight.ErrInternalf("failed to create session").WithError(err)
	}

	return writeJSON(w, http.StatusOK, Response{
		Status:  StatusOK,
		Message: "login successfull",
		Data: W{
			"session": session,
		},
	})
}
