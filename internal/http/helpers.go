package http

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	echosight "github.com/alexjoedt/echosight/internal"
	"github.com/alexjoedt/echosight/internal/validator"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func ReadURLParam(r *http.Request, key string) (string, error) {
	id := chi.URLParam(r, key)
	if id == "" {
		return "", echosight.ErrInvalidf("no '%s' param in URL", key)
	}
	return id, nil
}

func ReadUUIDParam(r *http.Request, key string) (uuid.UUID, error) {
	if key == "" {
		key = "id"
	}

	id := chi.URLParam(r, key)
	if id == "" {
		return uuid.Nil, echosight.ErrInvalidf("no '%s' param in URL", key)
	}

	u, err := uuid.Parse(id)
	if err != nil {
		return uuid.Nil, echosight.ErrInvalidf("param '%s' is not a valid uuid", key)
	}

	return u, nil
}

func ReadCSV(qs url.Values, key string, defaulValue []string) []string {
	csv := qs.Get(key)
	if csv == "" {
		return defaulValue
	}

	return strings.Split(csv, ",")
}

func ReadString(qs url.Values, key string, defaulValue string) string {
	s := qs.Get(key)
	if s == "" {
		return defaulValue
	}

	return s
}

func ReadTime(qs url.Values, key string, defaulValue time.Time) time.Time {
	s := qs.Get(key)
	if s == "" {
		return defaulValue
	}
	t, err := time.Parse(time.DateOnly, s)
	if err != nil {
		return defaulValue
	}
	return t
}

func ReadTimeString(qs url.Values, key string, defaulValue string) string {
	s := qs.Get(key)
	if s == "" {
		return defaulValue
	}
	t, err := time.Parse(time.DateOnly, s)
	if err != nil {
		return defaulValue
	}
	return t.Format(time.DateOnly)
}

func ReadInt(qs url.Values, key string, defaultValue int, v *validator.Validator) int {
	s := qs.Get(key)
	if s == "" {
		return defaultValue
	}

	i, err := strconv.Atoi(s)
	if err != nil {
		v.AddError(key, "must be an interger value")
		return defaultValue
	}

	return i
}

// ReadBoolParam reads the key from the url query and returns true if
// the param is equal 'true' otherwise it returns false.
func ReadBool(qs url.Values, key string) bool {
	return qs.Get(key) == "true"
}
