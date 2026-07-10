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
	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
)

type App struct {
	cfg    *config.Config
	server *http.Server
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

	router := mux.NewRouter().PathPrefix("/api/v1").Subrouter()

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

	tasks.NewModule(db, router)
	team.NewModule(db, router)
	auth.NewModule(db, client, router, cfg.Auth)

	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.Port),
		Handler:           router,
		ReadTimeout:       15 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
		MaxHeaderBytes:    1 << 20, // 1 МБ
	}

	return &App{cfg: cfg, server: server}, nil
}

func (a *App) Run(ctx context.Context) {

	go func() {
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	<-ctx.Done()

	shutCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := a.server.Shutdown(shutCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
}
