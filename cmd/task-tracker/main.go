package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/KronusRodion/task-tracker/internal/app"
	"github.com/KronusRodion/task-tracker/internal/config"
	"github.com/KronusRodion/task-tracker/internal/constants"
)

func main() {
	path, ok := os.LookupEnv(constants.ConfPath)
	if !ok || path == "" {
		log.Fatalf("%s is not set", constants.ConfPath)
	}

	cfg, err := config.Load(path)
	if err != nil {
		log.Fatal(err)
	}


	app, err := app.Build(cfg)
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	go func ()  {
		select {
		case <-ctx.Done():
		case sig := <-sigChan:
			log.Println("Завершение по сигналу: ", sig)
			cancel()
		}	
	}()
	

	app.Run(ctx)
}
