package http

import (
	"net/http"

	echosight "github.com/alexjoedt/echosight/internal"
	"github.com/alexjoedt/echosight/internal/filter"
	"github.com/alexjoedt/echosight/internal/validator"
	"github.com/go-chi/chi/v5"
)

func (s *Server) registerPreferencesRoutes(r *chi.Mux) {

	r.With(s.requireAuth).Route("/preferences", func(r chi.Router) {
		r.Post("/", makeHandlerFunc(s.handlerCreatePreferences))
		r.Get("/", makeHandlerFunc(s.handlerListPreferences))
		r.Get("/{prefName}", makeHandlerFunc(s.handlerGetPreferenceByName))
	})
}

func (s *Server) handlerCreatePreferences(w http.ResponseWriter, r *http.Request) error {
	ctx, cancel := s.ServerContext(r.Context())
	defer cancel()

	var input echosight.Preferences

	err := readJSON(r, &input)
	if err != nil {
		return err
	}

	input.CryptValues(s.Crypter, "password", "key", "secret", "token")
	v := validator.New()
	input.Validate(v)

	if !v.Valid() {
		return echosight.ErrConflictf("invalid payload").WithData(v.Errors)
	}

	// TODO: is this a good idea to overwrite ?
	err = s.PreferenceService.SetAll(ctx, &input)
	if err != nil {
		return err
	}

	return writeJSON(w, http.StatusOK, Response{
		Status:  StatusOK,
		Message: "preferences created",
		Data: W{
			"preferences": input.Map(),
		},
	})
}

func (s *Server) handlerGetPreferenceByName(w http.ResponseWriter, r *http.Request) error {
	ctx, cancel := s.ServerContext(r.Context())
	defer cancel()

	prefName, err := ReadURLParam(r, "prefName")
	if err != nil {
		return err
	}

	pref, err := s.PreferenceService.GetByName(ctx, prefName)
	if err != nil {
		return err
	}

	return writeJSON(w, http.StatusOK, Response{
		Status: StatusOK,
		Data: W{
			prefName: pref.Value,
		},
	})
}

func (s *Server) handlerListPreferences(w http.ResponseWriter, r *http.Request) error {
	ctx, cancel := s.ServerContext(r.Context())
	defer cancel()

	qs := r.URL.Query()
	f := filter.NewDefaultPreferenceFilter()
	f.Name = ReadString(qs, "name", "")

	prefs, err := s.PreferenceService.List(ctx, f)
	if err != nil {
		return err
	}

	return writeJSON(w, http.StatusOK, Response{
		Status: StatusOK,
		Data: W{
			"preferences": prefs,
		},
	})
}
