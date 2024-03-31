package http

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	echosight "github.com/alexjoedt/echosight/internal"
	"github.com/alexjoedt/echosight/internal/logger"
	"github.com/rs/xid"
	"github.com/rs/zerolog"
	"golang.org/x/time/rate"
)

type loggingResponseWriter struct {
	w          http.ResponseWriter
	statusCode int
}

func newLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{w, http.StatusOK}
}

// Hijack implements http.Hijacker to work with websockets
func (lrw *loggingResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h, ok := lrw.w.(http.Hijacker)
	if !ok {
		return nil, nil, errors.New("hijack not supported")
	}
	return h.Hijack()
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.w.WriteHeader(code)
}

func (s *Server) loggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, _ := net.SplitHostPort(r.RemoteAddr)
		start := time.Now()
		lrw := newLoggingResponseWriter(w)

		correlationID := xid.New().String()

		ctx := echosight.NewContextWithCorrelationID(r.Context(), correlationID)
		s.log.UpdateContext(func(c zerolog.Context) zerolog.Context {
			return c.Str("correlation_id", correlationID)
		})
		w.Header().Add("X-Correlation-ID", correlationID)
		r = r.WithContext(s.log.WithContext(ctx))

		defer func() {
			msg := fmt.Sprintf("%s %d %s %.2fs", r.Method, lrw.statusCode, r.URL.RequestURI(), time.Since(start).Seconds())
			s.log.Debugw(msg, logger.Str("ip", ip), logger.Str("user_agent", r.UserAgent()))
		}()
		next.ServeHTTP(lrw.w, r)
	})
}

// reportPanic is middleware for catching panics and reporting them.
func (s *Server) recoverPanicMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				s.log.
					Err().
					Err(fmt.Errorf("%s", err)).
					Stack().
					Str("method", r.Method).
					Str("url", r.URL.RequestURI()).
					Msg("unexpected panic")

				w.Header().Set("Connection", "close")
				InternalServerError(w, "fatal error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func (s *Server) rateLimiterMiddleware(next http.Handler) http.Handler {
	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}

	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
	)

	go func() {
		for {
			time.Sleep(time.Minute)
			mu.Lock()
			for ip, client := range clients {
				if time.Since(client.lastSeen) > 3*time.Minute {
					delete(clients, ip)
				}
			}
			mu.Unlock()
		}
	}()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if s.RateLimiter.Enabled {
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				InternalServerError(w, "invalid remote")
				return
			}

			mu.Lock()
			if _, found := clients[ip]; !found {
				clients[ip] = &client{
					limiter: rate.NewLimiter(rate.Limit(s.RateLimiter.Limit), s.RateLimiter.Burst),
				}
			}

			clients[ip].lastSeen = time.Now()

			if !clients[ip].limiter.Allow() {
				mu.Unlock()
				TooManyRequestsError(w, "rate limit exceeded")
				return
			}

			mu.Unlock()
		}
		next.ServeHTTP(w, r)
	})
}
