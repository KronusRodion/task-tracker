package app

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/KronusRodion/task-tracker/internal/api/auth"
	"github.com/KronusRodion/task-tracker/internal/api/tasks"
	"github.com/KronusRodion/task-tracker/internal/api/team"
	"github.com/KronusRodion/task-tracker/internal/closer"
	"github.com/KronusRodion/task-tracker/internal/config"
	"github.com/KronusRodion/task-tracker/internal/metrics"
	"github.com/KronusRodion/task-tracker/internal/middleware"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
)

type App struct {
	cfg     *config.Config
	server  *http.Server
	metrics *metrics.Metrics
	closer  *closer.Closer
}

func Build(cfg *config.Config) (*App, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	closer := closer.New()
	var err error
	defer func() {
		if err != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_ = closer.Close(ctx)
		}
	}()

	db, err := sql.Open("mysql", cfg.Database.MySQLDSN())
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}

	closer.Add(db)

	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Cache.Address(),
		Password: cfg.Cache.Password,
		DB:       cfg.Cache.DB,
	})

	err = client.Ping(ctx).Err()
	if err != nil {
		return nil, err
	}
	closer.Add(client)

	m := metrics.NewMetrics()
	limiter := middleware.NewRateLimiter()

	mainRouter := mux.NewRouter()

	mainRouter.Handle("/metrics", promhttp.Handler())

	apiRouter := mainRouter.PathPrefix("/api/v1").Subrouter()

	apiRouter.Use(m.Middleware)
	apiRouter.Use(limiter.Middleware)

	auth.NewModule(db, client, apiRouter, cfg.Auth)
	tasks.NewModule(db, client, apiRouter)
	team.NewModule(db, apiRouter)

	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.Port),
		Handler:           mainRouter,
		ReadTimeout:       15 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
		MaxHeaderBytes:    1 << 20, // 1 МБ
	}

	return &App{
		cfg:     cfg,
		server:  server,
		metrics: m,
		closer:  closer,
	}, nil
}

func (a *App) Run(ctx context.Context) {
	go func() {
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	log.Println("app has started")
	<-ctx.Done()

	shutCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := a.server.Shutdown(shutCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
}

func (a *App) Close(ctx context.Context) {
	a.closer.Close(ctx)
}
