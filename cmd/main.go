package main

import (
	"log/slog"
	"os"

	//"github.com/numbermax/blockchain/internal/config"
	"github.com/numbermax/blockchain/cmd/cli"
	"github.com/numbermax/blockchain/internal/lib/logger/handlers/slogpretty"
	//"github.com/numbermax/blockchain/internal/services/wallet"
)

const (
	envLocal = "local"
	envDev   = "development"
	envProd  = "production"
)

func main() {

	defer os.Exit(0)
	// Load configuration
	// config := config.MustLoad()

	logger := setupLogger("local")

	// logger.Info(config.Env)

	cli := cli.CommandLine{Logger: logger}
	cli.Run()
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = setupPrettySlog()
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}

func setupPrettySlog() *slog.Logger {
	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}

	handler := opts.NewPrettyHandler(os.Stdout)

	return slog.New(handler)
}
