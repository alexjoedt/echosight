package http

import (
	"context"
	"net/http"
	"time"

	echosight "github.com/alexjoedt/echosight/internal"
	"github.com/google/uuid"
)

// handleNotFound handles requests to routes that don't exist.
func (s *Server) handleNotFound(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusNotFound, Response{
		Status:  StatusWarn,
		Message: "path not found",
	})
}

// handleNotImplemented handles requests to routes that should implemented
func (s *Server) handleNotImplemented(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusNotImplemented, Response{
		Status:  StatusWarn,
		Message: "route not implemented",
	})
}

// handleVersion displays the deployed version.
func (s *Server) handleVersion(w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(echosight.Version))
	return nil
}

// handleVersion displays the deployed commit.
func (s *Server) handleRevision(w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(echosight.Revision))
	return nil
}

// handleInfo displays app information.
func (s *Server) handleAppInfo(w http.ResponseWriter, r *http.Request) error {
	return writeJSON(w, http.StatusOK, W{
		"name":     echosight.Name,
		"version":  echosight.Version,
		"revision": echosight.Revision,
	})
}

// handleInfo displays app information.
func (s *Server) handleHealthCheck(w http.ResponseWriter, r *http.Request) error {
	status := "ok"
	message := ""
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	_, err := s.UserService.GetByID(ctx, uuid.MustParse(echosight.AdminUserID))
	if err != nil {
		if ctx.Err() != nil {
			status = "warn"
			message = "long db query duration"
		} else {
			status = "error"
			message = "database error"
		}
	}

	return writeJSON(w, http.StatusOK, W{
		"status":  status,
		"message": message,
	})
}
