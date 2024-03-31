package http

import (
	"net/http"

	echosight "github.com/alexjoedt/echosight/internal"
)

// authenticatedMiddleware adds a user to the context, if there is no session, it adds
// an anonymous user
func (s *Server) authenticatedMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Authorization")

		token, err := BearerAuth(r)
		if err != nil {
			r = r.WithContext(echosight.NewContextWithUser(r.Context(), echosight.AnonymusUser))
			next.ServeHTTP(w, r)
			return
		}

		session, _, err := s.SessionService.Get(r.Context(), token)
		if err != nil {
			s.log.Errorc("%v", err)
			InvalidSession(w)
			return
		}

		// Set the user to the context
		r = r.WithContext(echosight.NewContextWithUser(r.Context(), session.User))
		next.ServeHTTP(w, r)
	})
}

func (s *Server) requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, err := echosight.UserFromContext(r.Context())
		if err != nil || user.IsAnonymus() {
			switch {
			case user == nil:
				s.log.Debugf("requireAuth: user is nil")
			case user.IsAnonymus():
				s.log.Debugf("requireAuth: user is anonymous")
			}
			NotAuthenticaded(w)
			return
		}

		if !user.Activated {
			Forbidden(w, "your user account must be activated to access this resource")
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) requireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		user, err := echosight.UserFromContext(r.Context())
		if err != nil {
			s.log.Errorf("%v", err)
			BadRequest(w, "no user ID")
			return
		}

		if !user.IsAdmin() {
			Forbidden(w, "no access to this ressource")
			return
		}

		next.ServeHTTP(w, r)
	})
}
