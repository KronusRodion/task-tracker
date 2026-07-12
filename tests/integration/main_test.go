package integration

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/KronusRodion/task-tracker/internal/app"
	"github.com/KronusRodion/task-tracker/internal/config"
	migration "github.com/KronusRodion/task-tracker/internal/migrations"
	"github.com/KronusRodion/task-tracker/tests/integration/env"
	_ "github.com/go-sql-driver/mysql"
)

var TestEnv *env.Env

func TestMain(m *testing.M) {
	fmt.Println("=== TestMain started ===")
	ctx := context.Background()

	var err error
	TestEnv, err = env.New(ctx)
	if err != nil {
		fmt.Printf("=== Error initializing test environment: %v ===\n", err)
		os.Exit(1)
	}
	fmt.Println("=== Test environment initialized successfully ===")

	cfg := config.Default()
	host, err := TestEnv.MySQL.Host(ctx)
	if err != nil {
		fmt.Printf("=== Error looking db host: %v ===\n", err)
		os.Exit(1)
	}
	port, err := TestEnv.MySQL.MappedPort(ctx, "3306")
	if err != nil {
		fmt.Printf("=== Error looking db port: %v ===\n", err)
		os.Exit(1)
	}
	hostCache, err := TestEnv.Redis.Host(ctx)
	if err != nil {
		fmt.Printf("=== Error looking cache host: %v ===\n", err)
		os.Exit(1)
	}
	println()
	portCache, err := TestEnv.Redis.MappedPort(ctx, "6379")
	if err != nil {
		fmt.Printf("=== Error looking cache port: %v ===\n", err)
		os.Exit(1)
	}

	cfg.Database.Host = host
	cfg.Cache.Host = hostCache
	cfg.Database.Port = int(port.Num())
	cfg.Cache.Port = int(portCache.Num())
	err = cfg.Validate()
	if err != nil {
		fmt.Printf("=== Error validating cfg: %v ===\n", err)
		os.Exit(1)
	}

	app, err := app.Build(cfg)
	if err != nil {
		fmt.Printf("=== Error building app: %v ===\n", err)
		os.Exit(1)
	}

	go func() {
		app.Run(ctx)
	}()

	err = migration.ApplyMigrations(TestEnv.DB, "../../migrations")
	if err != nil {
		fmt.Printf("=== Error apply migrations: %v ===\n", err)
		os.Exit(1)
	}

	time.Sleep(5 * time.Second)

	code := m.Run()

	TestEnv.Close()
	fmt.Println("=== Test environment closed ===")
	os.Exit(code)
}
