/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"osom/cmd"
	config "osom/pkg"
	"syscall"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}

	config.InitConfig()
	config.InitLogging()
	config.InitMQTT()

	// Create a context that can be cancelled on SIGINT or SIGTERM. This overrides Go's default behavior of
	// terminating the program immediately on these signals, allowing for graceful shutdown and flushing of traces.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	// Setup OpenTelemetry SDK
	shutdown, err := config.SetupOTelSDK(ctx)
	if err != nil {
		panic(err)
	}

	err = sentry.Init(sentry.ClientOptions{
		Dsn:   os.Getenv("SENTRY_DSN"),
		Debug: true,
	})
	if err != nil {
		panic(err)
	}

	defer func() {
		sentry.Flush(2 * time.Second)
		fmt.Println("Shutting down OTel SDK...")
		if err := shutdown(context.Background()); err != nil {
			fmt.Printf("Error during OTel SDK shutdown: %v\n", err)
		}
		stop()
	}()

	cmd.ExecuteContext(ctx)
}
