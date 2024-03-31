package http

import (
	"net/http"
	"time"

	echosight "github.com/alexjoedt/echosight/internal"
	"github.com/alexjoedt/echosight/internal/filter"
	"github.com/alexjoedt/echosight/internal/validator"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func (s *Server) handlerCreateRecipient(w http.ResponseWriter, r *http.Request) error {
	ctx, cancel := s.ServerContext(r.Context())
	defer cancel()

	var input struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	err := readJSON(r, &input)
	if err != nil {
		s.log.Errorc("failed to read recipient payload", err)
		return err
	}

	recipient := echosight.Recipient{
		Name:  input.Name,
		Email: input.Email,
	}

	v := validator.New()
	ValidateRecipient(v, &recipient)
	if !v.Valid() {
		return echosight.ErrInvalidf("invalid recipient payload").WithData(v.Errors)
	}

	err = s.RecipientService.Create(ctx, &recipient)
	if err != nil {
		return err
	}

	return writeJSON(w, http.StatusOK, Response{
		Status:  StatusOK,
		Message: "recipient created",
		Data: W{
			"recipient": recipient,
		},
	})
}

func (s *Server) handlerGetRecipientByID(w http.ResponseWriter, r *http.Request) error {
	ctx, cancel := s.ServerContext(r.Context())
	defer cancel()

	id := chi.URLParam(r, "recipientID")
	if id == "" {
		return echosight.ErrInvalidf("no recipient ID provided")
	}

	recipientID, err := uuid.Parse(id)
	if err != nil {
		s.log.Errorc("invalid id: %v", err)
		return echosight.ErrInvalidf("invalid ID '%s'", id)
	}

	recipient, err := s.RecipientService.GetByID(ctx, recipientID)
	if err != nil {
		return err
	}

	return writeJSON(w, http.StatusOK, Response{
		Status: StatusOK,
		Data: W{
			"recipient": recipient,
		},
	})
}

func (s *Server) handlerGetRecipients(w http.ResponseWriter, r *http.Request) error {
	ctx, cancel := s.ServerContext(r.Context())
	defer cancel()

	// TODO: read filter params
	recipientFilter := filter.NewDefaultRecipientFilter()

	recipients, err := s.RecipientService.List(ctx, recipientFilter)
	if err != nil {
		return err
	}

	return writeJSON(w, http.StatusOK, Response{
		Status: StatusOK,
		Data: W{
			"recipients": recipients,
			"pagination": recipientFilter.Pagination,
		},
	})
}

func (s *Server) handlerUpdateRecipient(w http.ResponseWriter, r *http.Request) error {
	ctx, cancel := s.ServerContext(r.Context())
	defer cancel()

	id := chi.URLParam(r, "recipientID")
	if id == "" {
		return echosight.ErrInvalidf("no recipient ID provided")
	}

	recipientID, err := uuid.Parse(id)
	if err != nil {
		s.log.Errorc("invalid id: %v", err)
		return echosight.ErrInvalidf("invalid ID '%s'", id)
	}

	var input struct {
		Name      *string `json:"name"`
		Email     *string `json:"email"`
		Activated *bool   `json:"activated"`
	}

	err = readJSON(r, &input)
	if err != nil {
		s.log.Errorc("failed to read recipient payload", err)
		return err
	}

	recipient, err := s.RecipientService.GetByID(ctx, recipientID)
	if err != nil {
		return err
	}

	if input.Name != nil {
		recipient.Name = *input.Name
	}

	if input.Email != nil {
		recipient.Email = *input.Email
	}

	if input.Activated != nil {
		recipient.Activated = *input.Activated
	}

	recipient.UpdatedAt = time.Now()

	v := validator.New()
	ValidateRecipient(v, recipient)
	if !v.Valid() {
		return echosight.ErrInvalidf("invalid recipient payload").WithData(v.Errors)
	}

	err = s.RecipientService.Update(ctx, recipient)
	if err != nil {
		return err
	}

	return writeJSON(w, http.StatusOK, Response{
		Status:  StatusOK,
		Message: "recipient updated",
		Data: W{
			"recipient": recipient,
		},
	})
}

func (s *Server) handlerDeleteRecipientByID(w http.ResponseWriter, r *http.Request) error {
	ctx, cancel := s.ServerContext(r.Context())
	defer cancel()

	id := chi.URLParam(r, "recipientID")
	if id == "" {
		return echosight.ErrInvalidf("no recipient ID provided")
	}

	recipientID, err := uuid.Parse(id)
	if err != nil {
		s.log.Errorc("invalid id: %v", err)
		return echosight.ErrInvalidf("invalid ID '%s'", id)
	}

	recipient, err := s.RecipientService.DeleteByID(ctx, recipientID)
	if err != nil {
		return err
	}

	return writeJSON(w, http.StatusOK, Response{
		Status: StatusOK,
		Data: W{
			"recipient": recipient,
		},
	})
}

func ValidateRecipient(v *validator.Validator, recipient *echosight.Recipient) {
	v.Check(len(recipient.Name) > 3, "name", "name too short")
	v.Check(validator.IsEmail(recipient.Email), "email", "invalid email")
}
