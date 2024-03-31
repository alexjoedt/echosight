package http

import "github.com/go-chi/chi/v5"

func (s *Server) registerAuthRoutes(r *chi.Mux) {
	r.Post("/login", makeHandlerFunc(s.handleLogin))
}

func (s *Server) registerUserRoutes(r *chi.Mux) {
	r.With(s.requireAdmin).Route("/users", func(r chi.Router) {
		r.Get("/", s.handleNotImplemented)
	})
}

func (s *Server) registerRecipientRoutes(r *chi.Mux) {
	r.With(s.requireAdmin).Route("/recipients", func(r chi.Router) {
		r.Get("/", makeHandlerFunc(s.handlerGetRecipients))
		r.Get("/{recipientID}", makeHandlerFunc(s.handlerGetRecipientByID))
		r.Delete("/{recipientID}", makeHandlerFunc(s.handlerDeleteRecipientByID))
		r.Post("/", makeHandlerFunc(s.handlerCreateRecipient))
		r.Patch("/{recipientID}", makeHandlerFunc(s.handlerUpdateRecipient))
	})
}

func (s *Server) registerHostRoutes(r *chi.Mux) {
	r.With(s.requireAuth).Route("/hosts", func(r chi.Router) {
		r.Get("/", makeHandlerFunc(s.handlerGetHosts))
		r.Get("/{hostID}", makeHandlerFunc(s.handlerGetHostByID))
		r.Delete("/{hostID}", makeHandlerFunc(s.handlerDeleteHostByID))
		r.Post("/", makeHandlerFunc(s.handlerCreateHost))
		r.Patch("/{hostID}", makeHandlerFunc(s.handlerUpdateHost))
	})
}

// registerDetectorRoutes is a helper function to register routes to a mux.
// detectors belongs always to a host
func (s *Server) registerDetectorRoutes(r *chi.Mux) {
	r.With(s.requireAuth).Route("/hosts/{hostID}/detectors", func(r chi.Router) {
		r.Get("/", makeHandlerFunc(s.handlerGetDetectors))
		r.Post("/", makeHandlerFunc(s.handlerCreateDetector))

		r.Get("/{detectorID}", makeHandlerFunc(s.handlerGetDetectorByID))
		r.Patch("/{detectorID}", makeHandlerFunc(s.handlerUpdateDetector))
		r.Delete("/{detectorID}", makeHandlerFunc(s.handlerDeleteDetectorByID))

		r.Post("/{detectorID}/activate", makeHandlerFunc(s.handlerActivateDetector))
		r.Post("/{detectorID}/deactivate", makeHandlerFunc(s.handlerDeactivateDetector))
	})
}

// registerObserverRoutes
func (s *Server) registerObserverRoutes(r *chi.Mux) {
	r.With(s.requireAdmin).Route("/observer", func(r chi.Router) {
		// TODO: better name for route
		r.Post("/", makeHandlerFunc(s.handlerObserverStatus))
	})
}
