package main

import (
	"context"
	"database/sql"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog/log"

	"pipo-edu-project/internal/auth"
	"pipo-edu-project/internal/config"
	httpapi "pipo-edu-project/internal/http"
	"pipo-edu-project/internal/logging"
	"pipo-edu-project/internal/metrics"
	"pipo-edu-project/internal/repository"
	repo "pipo-edu-project/internal/repository/sqlc"
	"pipo-edu-project/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}
	logging.Init(cfg.Env)

	if cfg.MigrateOnStart {
		if err := repository.RunMigrations(cfg.DBDSN, "db/migrations"); err != nil {
			log.Fatal().Err(err).Msg("migrations failed")
		}
	}

	db, err := repository.Open(cfg.DBDSN)
	if err != nil {
		log.Fatal().Err(err).Msg("db connection failed")
	}
	defer db.Close()

	queries := repo.New(db)
	svc := service.New(queries)

	if cfg.BootstrapEmail != "" && cfg.BootstrapPassword != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		var count int
		if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users").Scan(&count); err != nil {
			log.Error().Err(err).Msg("bootstrap check failed")
		} else if count == 0 {
			hash, err := auth.HashPassword(cfg.BootstrapPassword)
			if err != nil {
				log.Error().Err(err).Msg("bootstrap hash failed")
			} else {
				_, err := queries.CreateUser(ctx, repo.CreateUserParams{
					Email:        cfg.BootstrapEmail,
					PasswordHash: hash,
					Role:         string(auth.RoleAdmin),
					FullName:     cfg.BootstrapName,
					PlotNumber:   sql.NullString{},
					CreatedBy:    uuid.NullUUID{},
					UpdatedBy:    uuid.NullUUID{},
				})
				if err != nil {
					log.Error().Err(err).Msg("bootstrap user create failed")
				} else {
					log.Info().Msg("bootstrap admin created")
				}
			}
		}
	}

	registry := prometheus.NewRegistry()
	registry.MustRegister(prometheus.NewGoCollector())
	registry.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
	metricsCollector := metrics.New(registry)

	tokens := auth.NewTokenManager(cfg.JWTSecret, cfg.JWTRefreshSecret, cfg.AccessTTL, cfg.RefreshTTL)

	handler := &httpapi.Handler{
		Auth:    tokens,
		Service: svc,
		Metrics: metricsCollector,
		CORS:    cfg.CORSOrigins,
	}

	router := httpapi.NewRouter(handler)
	server := &http.Server{
		Addr:         cfg.HTTPAddr,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Info().Str("addr", cfg.HTTPAddr).Msg("server started")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("server error")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info().Msg("shutdown initiated")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("shutdown error")
	}
}
