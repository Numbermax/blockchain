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
	logger *slog.Logger
}

func (cli *ComandLine) printUsage() {
	fmt.Println("Usage:")
	fmt.Println(" getbalance -address ADDRESS - get balance for the address")
	fmt.Println(" createblockchain -address ADDRESS - created a blockchain")
	fmt.Println(" printchain - Prints the blocks in the chain")
	fmt.Println(" send -from FROM -to TO -amount AMOUNT - send amount to the address")
}

func (cli *ComandLine) ValidateArguments() {
	if len(os.Args) < 2 {
		cli.gracefullExit()
	}
}

func (cli *ComandLine) printChain() {
	chain := blockchain.ContinueBlockChain(*cli.logger, "")
	defer chain.Database.Close()
	iter := chain.Iterator()

	for {
		block := iter.Next()
		cli.logger.Info("Block", slog.String("hash", fmt.Sprintf("%x", block.Hash)))
		pow := blockchain.NewProof(block)
		cli.logger.Info("Proof of Work", slog.String("valid", strconv.FormatBool(pow.Validate())))
		if len(block.PrevHash) == 0 {
			break
		}
	}
}

func (cli *ComandLine) getBalance(address string) {
	chain := blockchain.ContinueBlockChain(*cli.logger, address)
	defer chain.Database.Close()

	balance := 0
	UTXOs := chain.FindUTXO(address)

	for _, out := range UTXOs {
		balance += out.Value
	}

	cli.logger.Info("Balance: counted", slog.String("Balance: ", fmt.Sprintf("%d", balance)), slog.String("address: ", address))
}

func (cli *ComandLine) send(from, to string, amount int) {
	chain := blockchain.ContinueBlockChain(*cli.logger, from)
	defer chain.Database.Close()

	tx := blockchain.NewTransaction(from, to, amount, chain)

	chain.AddBlock([]*blockchain.Transaction{tx})
	cli.logger.Info("Success")
}

func (cli *ComandLine) createBlockchain(address string) {
	chain := blockchain.InitBlockChain(*cli.logger, address)
	chain.Database.Close()
	cli.logger.Info("Finished")
}

func (cli *ComandLine) gracefullExit() {
	cli.printUsage()
	runtime.Goexit()
}

func (cli *ComandLine) run() {
	cli.ValidateArguments()

	getbalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)
	createblockchainCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("print", flag.ExitOnError)

	getBalanceAddress := getbalanceCmd.String("address", "", "Address to get balance for")
	createBlockChainAddress := createblockchainCmd.String("address", "", "Address to create blockchain for")
	sendFrom := sendCmd.String("from", "", "Address to send from")
	sendTo := sendCmd.String("to", "", "Address to send to")
	sendAmount := sendCmd.Int("amount", 0, "Amount to send")

	switch os.Args[1] {
	case "getbalance":
		err := getbalanceCmd.Parse(os.Args[2:])
		blockchain.ErrHandle(err)

	case "createblockchain":
		err := createblockchainCmd.Parse(os.Args[2:])
		blockchain.ErrHandle(err)

	case "send":
		err := sendCmd.Parse(os.Args[2:])
		blockchain.ErrHandle(err)

	case "printchain":
		err := printChainCmd.Parse(os.Args[2:])
		blockchain.ErrHandle(err)

	default:
		cli.gracefullExit()
	}

	if getbalanceCmd.Parsed() {
		if *getBalanceAddress == "" {
			cli.logger.Error("Address is required for getbalance command")
			cli.gracefullExit()
		}
		cli.getBalance(*getBalanceAddress)
	}
	if createblockchainCmd.Parsed() {
		if *createBlockChainAddress == "" {
			cli.logger.Error("Address is required for createblockchain command")
			cli.gracefullExit()
		}
		cli.createBlockchain(*createBlockChainAddress)
	}

	if sendCmd.Parsed() {
		if *sendFrom == "" || *sendTo == "" || *sendAmount <= 0 {
			cli.logger.Error("From, To and Amount are required for send command")
			cli.gracefullExit()
		}
		cli.send(*sendFrom, *sendTo, *sendAmount)
	}

	// Print chain
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

	cli := ComandLine{logger}
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
