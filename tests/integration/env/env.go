// tests/integration/env/env.go
package env

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"

	"github.com/testcontainers/testcontainers-go"
	tcmysql "github.com/testcontainers/testcontainers-go/modules/mysql"
	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"
)

type Env struct {
	ctx context.Context

	MySQL *tcmysql.MySQLContainer
	Redis *tcredis.RedisContainer

	DB          *sql.DB
	RedisClient *redis.Client
}

func New(ctx context.Context) (*Env, error) {
	e := &Env{
		ctx: ctx,
	}
	log.Println("starting env")
	if err := e.startMySQL(); err != nil {
		return nil, err
	}

	if err := e.startRedis(); err != nil {
		e.Close()
		return nil, err
	}

	return e, nil
}

func (e *Env) Close() {
	if e.DB != nil {
		e.DB.Close()
	}

	if e.RedisClient != nil {
		_ = e.RedisClient.Close()
	}

	if e.MySQL != nil {
		_ = e.MySQL.Terminate(e.ctx)
	}

	if e.Redis != nil {
		_ = e.Redis.Terminate(e.ctx)
	}
}

func (e *Env) startMySQL() error {
	container, err := tcmysql.Run(
		e.ctx,
		"mysql:8.4",
		tcmysql.WithDatabase("task_tracker"),
		tcmysql.WithUsername("root"),
		tcmysql.WithPassword("password"),
		testcontainers.WithWaitStrategy(
			wait.ForListeningPort("3306/tcp"),
		),
	)
	if err != nil {
		return err
	}

	e.MySQL = container

	dsn, err := container.ConnectionString(e.ctx)
	if err != nil {
		return err
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return err
	}

	if err := db.Ping(); err != nil {
		return err
	}

	e.DB = db
	return nil
}

func (e *Env) startRedis() error {
	container, err := tcredis.Run(
		e.ctx,
		"redis:7-alpine",
	)
	if err != nil {
		return err
	}

	e.Redis = container

	host, err := container.Host(e.ctx)
	if err != nil {
		return err
	}

	port, err := container.MappedPort(e.ctx, "6379")
	if err != nil {
		return err
	}

	addr := fmt.Sprintf("%s:%s", host, port.Port())

	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	if err := client.Ping(e.ctx).Err(); err != nil {
		return err
	}

	e.RedisClient = client

	return nil
}
