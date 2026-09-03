// Package httpserver implements HTTP server.
package httpserver

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/evrone/go-clean-template/pkg/logger"
	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"
)

const (
	_defaultAddr            = ":80"
	_defaultReadTimeout     = 5 * time.Second
	_defaultWriteTimeout    = 5 * time.Second
	_defaultShutdownTimeout = 3 * time.Second
)

// Server -.
type Server struct {
	ctx context.Context
	eg  *errgroup.Group

	App    *gin.Engine
	server *http.Server
	notify chan error

	address         string
	readTimeout     time.Duration
	writeTimeout    time.Duration
	shutdownTimeout time.Duration

	logger logger.Interface
}

// New -.
func New(l logger.Interface, opts ...Option) *Server {
	group, ctx := errgroup.WithContext(context.Background())
	group.SetLimit(1) // Run only one goroutine

	s := &Server{
		ctx:             ctx,
		eg:              group,
		App:             nil,
		server:          nil,
		notify:          make(chan error, 1),
		address:         _defaultAddr,
		readTimeout:     _defaultReadTimeout,
		writeTimeout:    _defaultWriteTimeout,
		shutdownTimeout: _defaultShutdownTimeout,
		logger:          l,
	}

	// Custom options
	for _, opt := range opts {
		opt(s)
	}

	gin.SetMode(gin.ReleaseMode)

	app := gin.New()

	s.App = app
	s.server = &http.Server{
		Addr:              s.address,
		Handler:           app,
		ReadTimeout:       s.readTimeout,
		ReadHeaderTimeout: s.readTimeout,
		WriteTimeout:      s.writeTimeout,
	}

	return s
}

// Start -.
func (s *Server) Start() {
	s.eg.Go(func() error {
		err := s.server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.notify <- err

			close(s.notify)

			return err
		}

		return nil
	})

	s.logger.Info("restapi server - Server - Started")
}

// Notify -.
func (s *Server) Notify() <-chan error {
	return s.notify
}

// Shutdown -.
func (s *Server) Shutdown() error {
	var shutdownErrors []error

	ctx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
	defer cancel()

	err := s.server.Shutdown(ctx)
	if err != nil && !errors.Is(err, context.Canceled) {
		s.logger.Error(err, "restapi server - Server - Shutdown - s.server.Shutdown")

		shutdownErrors = append(shutdownErrors, err)
	}

	// Wait for all goroutines to finish and get any error
	err = s.eg.Wait()
	if err != nil && !errors.Is(err, context.Canceled) {
		s.logger.Error(err, "restapi server - Server - Shutdown - s.eg.Wait")

		shutdownErrors = append(shutdownErrors, err)
	}

	s.logger.Info("restapi server - Server - Shutdown")

	return errors.Join(shutdownErrors...)
}
