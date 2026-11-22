package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/clodoaldomarques/core-sdk/pkg/otel/tracer"
	"github.com/clodoaldomarques/core-sdk/pkg/zap/logger"

	"github.com/clodoaldomarques/ledger-worker/internal/infra/rest/shared"
	"github.com/go-playground/validator"
	"github.com/labstack/echo/v4"
)

// HealthChecker é implementado pelo consumidor SQS usado na readiness.
type HealthChecker interface {
	Health() error
}

type Server struct {
	http   *echo.Echo
	health HealthChecker
}

func New(h HealthChecker) *Server {
	s := Server{
		http:   echo.New(),
		health: h,
	}
	s.routes()
	return &s
}

func (s Server) routes() {
	s.http.Validator = &CustomValidator{validator: validator.New()}

	// tracer interceptor
	s.http.Use(tracer.Interceptor())

	// logger interceptor
	s.http.Use(logger.Interceptor())

	// Health
	s.http.GET("/health/live", s.Liveness)
	s.http.GET("/health/ready", s.Readiness)
}

func (s Server) Start(port int) error {
	return s.http.Start(fmt.Sprintf(":%d", port))
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.http.Shutdown(ctx)
}

func (s *Server) Liveness(c echo.Context) error {
	return c.NoContent(http.StatusOK)
}

func (s *Server) Readiness(c echo.Context) error {
	ctx, cancel := context.WithTimeout(
		c.Request().Context(),
		500*time.Millisecond,
	)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- s.health.Health()
	}()

	select {
	case err := <-done:
		if err != nil {
			return c.NoContent(http.StatusServiceUnavailable)
		}
	case <-ctx.Done():
		return c.NoContent(http.StatusServiceUnavailable)
	}

	return c.NoContent(http.StatusOK)
}

type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		return err
	}

	r, ok := i.(shared.EntityRequest)
	if !ok {
		return nil
	}

	if err := r.Validate(); err != nil {
		return err
	}
	return nil
}
