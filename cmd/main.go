package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"strconv"

	// "github.com/numbermax/blockchain/internal/config"
	"github.com/numbermax/blockchain/internal/lib/logger/handlers/slogpretty"
	"github.com/numbermax/blockchain/internal/services/blockchain"
)

const (
	envLocal = "local"
	envDev   = "development"
	envProd  = "production"
)

type ComandLine struct {
	logger    *slog.Logger
	blockcain *blockchain.BlockChain
}

func (cli *ComandLine) printUsage() {
	fmt.Println("Usage:")
	fmt.Println(" add -block BLOCK_DATA - add a block to the chain")
	fmt.Println(" print - Prints the blocks in the chain")
}

func (cli *ComandLine) ValidateArguments() {
	if len(os.Args) < 2 {
		cli.gracefullExit()
	}
}

func (cli *ComandLine) addBlock(data string) {
	cli.blockcain.AddBlock(data)
	cli.logger.Info("Added Block")
}

func (cli *ComandLine) printChain() {
	iter := cli.blockcain.Iterator()

	for {
		block := iter.Next()

		cli.logger.Info("Block", slog.String("hash", fmt.Sprintf("%x", block.Hash)), slog.String("data", string(block.Data)))

		pow := blockchain.NewProof(block)
		cli.logger.Info("Proof of Work", slog.String("valid", strconv.FormatBool(pow.Validate())))

		if len(block.PrevHash) == 0 {
			break
		}
	}
}

func (cli *ComandLine) gracefullExit() {
	cli.printUsage()
	runtime.Goexit()
}

func (cli *ComandLine) run() {
	cli.ValidateArguments()

	addBlockCmd := flag.NewFlagSet("add", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("print", flag.ExitOnError)
	addBlockData := addBlockCmd.String("block", "", "BlockData")

	switch os.Args[1] {
	case "add":
		err := addBlockCmd.Parse(os.Args[2:])
		blockchain.ErrHandle(err)

	case "print":
		err := printChainCmd.Parse(os.Args[2:])
		blockchain.ErrHandle(err)

	default:
		cli.gracefullExit()
	}

	if addBlockCmd.Parsed() {
		if *addBlockData == "" {
			cli.gracefullExit()
		}
		cli.addBlock(*addBlockData)
	}

	if printChainCmd.Parsed() {
		cli.printChain()
	}

}

func main() {

	defer os.Exit(0)
	// Load configuration
	// config := config.MustLoad()

	logger := setupLogger("local")

	// logger.Info(config.Env)

	chain := blockchain.InitBlockChain()
	defer chain.Database.Close()

	cli := ComandLine{logger, chain}
	cli.run()
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
