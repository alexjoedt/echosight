package http

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	echosight "github.com/alexjoedt/echosight/internal"
	"github.com/alexjoedt/echosight/internal/eventflow"
	"github.com/alexjoedt/echosight/internal/logger"
	"github.com/alexjoedt/echosight/internal/observer"
	"github.com/go-chi/chi/v5"
	"github.com/rs/cors"
)

type RateLimiter struct {
	Enabled bool
	Burst   int
	Limit   int
}

type ServerOpts struct {
	TrustedOrigins []string
	Addr           string
}

// Server is a simple wrapper for e.g. chi mux to use custom hanlders wich returns an error
type Server struct {
	server *http.Server
	mux    *chi.Mux
	log    *logger.Logger
	IsDev  bool

	addr           string
	trustedOrigins []string

	sessionLifetime time.Duration

	// TODO: TLS

	RateLimiter RateLimiter

	UserService       echosight.UserService
	HostService       echosight.HostService
	DetectorService   echosight.DetectorService
	RecipientService  echosight.RecipientService
	PreferenceService echosight.PreferenceService
	SessionService    echosight.SessionService

	MetricReader echosight.MetricReader
	Scheduler    *observer.Scheduler
	EventHandler *eventflow.Engine
	Crypter      echosight.Crypter
}

// NewServer creates a new EchoSight server with
// all routes mapped
func NewServer(opts ServerOpts) (*Server, error) {
	addr := opts.Addr
	if !strings.Contains(addr, ":") {
		addr = fmt.Sprintf(":%s", addr)
	}

	s := &Server{
		addr:            addr,
		trustedOrigins:  opts.TrustedOrigins,
		mux:             chi.NewMux(),
		log:             logger.New("server"),
		sessionLifetime: time.Hour * 48,
	}

	// WebSockets
	wsManager := NewWebSocketManager(s)
	wsManager.TrustedOrigins = s.trustedOrigins

	s.mux.NotFound(s.handleNotFound)

	// WebSocket Router needs the Hijack Method implemented on middleware
	// https://stackoverflow.com/a/63096684
	eventRouter := chi.NewRouter()
	eventRouter.Use(s.loggerMiddleware)
	eventRouter.Get("/", wsManager.serveWS) // WebSockets

	// Register API routes
	apiV1Router := chi.NewRouter()
	apiV1Router.Use(s.recoverPanicMiddleware)
	apiV1Router.Use(s.loggerMiddleware) // sets the correlation_id
	apiV1Router.Use(s.rateLimiterMiddleware)
	apiV1Router.Use(s.authenticatedMiddleware) // adds a user or an anonymous user to the context

	apiV1Router.Get("/debug/version", makeHandlerFunc(s.handleVersion))
	apiV1Router.Get("/debug/revision", makeHandlerFunc(s.handleRevision))
	apiV1Router.Get("/health", makeHandlerFunc(s.handleHealthCheck))
	apiV1Router.Get("/info", makeHandlerFunc(s.handleAppInfo))

	// auth Routes
	s.registerAuthRoutes(apiV1Router)

	// Email recipients
	s.registerRecipientRoutes(apiV1Router)

	// user routes
	s.registerUserRoutes(apiV1Router)

	// host routes
	s.registerHostRoutes(apiV1Router)

	// detector routes
	s.registerDetectorRoutes(apiV1Router)

	// observer routes, these routes used to control the observer scheduler
	s.registerObserverRoutes(apiV1Router)

	// preference routes
	s.registerPreferencesRoutes(apiV1Router)

	s.mux.Mount("/api/v1", apiV1Router)

	// WebSocket Router
	s.mux.Mount("/events", eventRouter)
	return s, nil
}

// Run starts the server and blocks the code
func (s *Server) Run() error {

	// Set up cors
	c := cors.New(cors.Options{
		AllowedOrigins:   s.trustedOrigins,
		AllowCredentials: true,
		AllowedMethods:   []string{"GET", "POST", "PATCH", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"X-Client", "Content-Type", "Access-Control-Allow-Credentials", "Authorization"},
		Debug:            s.IsDev,
	})

	corsHandler := c.Handler
	s.server = &http.Server{
		Addr:    s.addr,
		Handler: corsHandler(s.mux),
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGTERM, os.Interrupt)
	go func() { s.server.ListenAndServe() }()

	<-done
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	err := s.server.Shutdown(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

type HandlerFunc func(w http.ResponseWriter, r *http.Request) error

// makeHandlerFunc decorates the HandlerFunc
func makeHandlerFunc(f HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := f(w, r)
		if err != nil {
			// get application error
			appError := echosight.FromError(err)

			// no logs here, logging should be in the handler to get the correct caller
			response := Response{
				Status:  StatusErr,
				Message: appError.Message,
				Errors:  appError.Data,
			}

			if appError.Data != nil {
				response.Errors = appError.Data
			}

			writeJSON(w, ErrorStatusCode(appError.Code), response)
		}
	}
}
