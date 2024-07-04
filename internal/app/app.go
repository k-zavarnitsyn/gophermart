package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/k-zavarnitsyn/gophermart/internal/api"
	"github.com/k-zavarnitsyn/gophermart/internal/config"
	"github.com/k-zavarnitsyn/gophermart/internal/container"
	"github.com/k-zavarnitsyn/gophermart/internal/middleware"
	log "github.com/sirupsen/logrus"
)

type ServerApp struct {
	cfg *config.Config
	cnt *container.Container

	saveTicker *time.Ticker
}

func NewServerApp(cfg *config.Config, cnt *container.Container) *ServerApp {
	return &ServerApp{
		cfg: cfg,
		cnt: cnt,
	}
}

func (s *ServerApp) Run(ctx context.Context) {
	log.SetLevel(s.cfg.Log.Level)
	log.SetReportCaller(s.cfg.Log.WithReportCaller)

	if s.cfg.UseDB() {
		defined, err := s.cnt.SchemaCreator().SchemaDefined(ctx)
		if err != nil {
			log.WithError(err).Fatal("failed to check DB schema")
		}
		if !defined {
			if err := s.cnt.SchemaCreator().CreateSchema(ctx); err != nil {
				log.WithError(err).Fatal("failed to create DB schema")
			}
		}
	} else {
		log.Fatal("DB is required")
	}

	serverAPI := api.New(
		s.cfg,
		s.cnt.Auth(),
		s.cnt.Gophermart(),
		s.cnt.Pinger(),
	)
	router := NewRouter(s.cnt)
	router.Use(middleware.WithGzipRequest)
	router.Use(middleware.WithGzipResponse)
	logger := middleware.NewLogger(&s.cfg.Log)
	router.Use(logger.WithRequestLogging, logger.WithResponseLogging)
	router.InitRoutes(serverAPI, true)

	server := &http.Server{
		Addr:              s.cfg.Address,
		Handler:           router,
		ReadHeaderTimeout: s.cfg.Server.ReadHeaderTimeout,
	}
	fmt.Printf("Starting server with config: %+v\n", s.cfg)
	go func() {
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Errorf("server starting error: %v", err)
		}
		log.Println("Stopped serving new connections.")
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	fmt.Println("Shutting down server...")
	shutdownCtx, shutdownRelease := context.WithTimeout(context.Background(), s.cfg.Server.ShutdownTimeout)
	defer shutdownRelease()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Errorf("shutdown error: %v", err)
	}
	if err := s.cnt.Shutdown(shutdownCtx); err != nil {
		log.Errorf("container shutdown error: %v", err)
	}

	fmt.Println("Shutdown complete")
}
