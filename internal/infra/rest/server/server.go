package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/clodoaldomarques/core-sdk/pkg/otel/tracer"
	"github.com/clodoaldomarques/core-sdk/pkg/zap/logger"

	"github.com/clodoaldomarques/ledger-worker/internal/infra/message"
	"github.com/clodoaldomarques/ledger-worker/internal/infra/rest/shared"
	"github.com/go-playground/validator"
	"github.com/labstack/echo/v4"
)

type Server struct {
	http *echo.Echo
}

func New() *Server {
	s := Server{
		http: echo.New(),
	}
	s.routes()
	return &s
}

func (s Server) routes() {
	s.http.Validator = &CustomValidator{validator: validator.New()}

	// tracer interceptor
	s.http.Use(tracer.Interceptor())

	// logger interceptor
	s.http.Use(logger.InterceptorWithConfig(logger.InterceptorConfig{
		MaxBodySize:     5 * 1024,
		LogRequestBody:  true,
		LogResponseBody: false,
		RedactFields:    []string{"password", "token", "credit_card"},
		SkipPaths: []string{
			"/health/live",
			"/health/ready",
		},
	}))

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

	if err := message.NewConsumer(ctx).Health(); err != nil {
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
