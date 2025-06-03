package main

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"

	"github.com/numbermax/blockchain/internal/config"
	"github.com/numbermax/blockchain/internal/lib/logger/handlers/slogpretty"
	"github.com/numbermax/blockchain/internal/services/blockchain"
)

const (
	envLocal = "local"
	envDev   = "development"
	envProd  = "production"
)

func main() {

	// Load configuration
	config := config.MustLoad()

	logger := setupLogger(config.Env)

	logger.Info(config.Env)

	chain := blockchain.InitBlockChain()
	chain.AddBlock("First block after genesis")
	chain.AddBlock("Second block after genesis")
	chain.AddBlock("Third block after genesis")

	for _, block := range chain.GetBlocks() {
		logger.Info("Block", slog.String("hash", fmt.Sprintf("%x", block.Hash)), slog.String("data", string(block.Data)))

		pow := blockchain.NewProof(block)
		logger.Info("Proof of Work", slog.String("valid", strconv.FormatBool(pow.Validate())))
	}
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
