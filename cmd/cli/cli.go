package cli

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"strconv"

	"github.com/numbermax/blockchain/internal/services/blockchain"
	"github.com/numbermax/blockchain/internal/services/wallet"
)

type CommandLine struct {
	Logger *slog.Logger
}

func (cli *CommandLine) printUsage() {
	fmt.Println("Usage:")
	fmt.Println(" getbalance -address ADDRESS - get balance for the address")
	fmt.Println(" createblockchain -address ADDRESS - created a blockchain")
	fmt.Println(" printchain - Prints the blocks in the chain")
	fmt.Println(" send -from FROM -to TO -amount AMOUNT - send amount to the address")
	fmt.Println(" createwallet - Creates a new wallet")
	fmt.Println(" listaddresses - Lists all addresses in the wallet")
}

func (cli *CommandLine) ValidateArguments() {
	if len(os.Args) < 2 {
		cli.gracefullExit()
	}
}

func (cli *CommandLine) printChain() {
	chain := blockchain.ContinueBlockChain(*cli.Logger, "")
	defer chain.Database.Close()
	iter := chain.Iterator()

	for {
		block := iter.Next()
		cli.Logger.Info("Block", slog.String("hash", fmt.Sprintf("%x", block.Hash)))
		pow := blockchain.NewProof(block)
		cli.Logger.Info("Proof of Work", slog.String("valid", strconv.FormatBool(pow.Validate())))
		if len(block.PrevHash) == 0 {
			break
		}
	}
}

func (cli *CommandLine) getBalance(address string) {
	chain := blockchain.ContinueBlockChain(*cli.Logger, address)
	defer chain.Database.Close()

	balance := 0
	UTXOs := chain.FindUTXO(address)

	for _, out := range UTXOs {
		balance += out.Value
	}

	cli.Logger.Info("Balance: counted", slog.String("Balance: ", fmt.Sprintf("%d", balance)), slog.String("address: ", address))
}

func (cli *CommandLine) send(from, to string, amount int) {
	chain := blockchain.ContinueBlockChain(*cli.Logger, from)
	defer chain.Database.Close()

	tx := blockchain.NewTransaction(from, to, amount, chain)

	chain.AddBlock([]*blockchain.Transaction{tx})
	cli.Logger.Info("Success")
}

func (cli *CommandLine) createBlockchain(address string) {
	chain := blockchain.InitBlockChain(*cli.Logger, address)
	chain.Database.Close()
	cli.Logger.Info("Finished")
}

func (cli *CommandLine) createWallet() {
	wallets, err := wallet.CreateWallets()
	if err != nil {
		cli.Logger.Error("Error creating wallets", slog.String("error", err.Error()))
	}
	address := wallets.AddWallet()
	cli.Logger.Info("New wallet created", slog.String("address", address))
	wallets.SaveFile()
	cli.Logger.Info("Wallets saved successfully")
	cli.Logger.Info("Finished")
}

func (cli *CommandLine) listAddresses() {
	wallets, err := wallet.CreateWallets()
	if err != nil {
		cli.Logger.Error("Error creating wallets", slog.String("error", err.Error()))
		cli.gracefullExit()
	}

	addresses := wallets.GetAllAddresses()
	if len(addresses) == 0 {
		cli.Logger.Info("No addresses found")
		return
	}

	for _, address := range addresses {
		cli.Logger.Info("Address", slog.String("address", address))
	}
}

func (cli *CommandLine) gracefullExit() {
	cli.printUsage()
	runtime.Goexit()
}

func (cli *CommandLine) Run() {
	cli.ValidateArguments()

	getbalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)
	createblockchainCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("print", flag.ExitOnError)
	createWalletCmd := flag.NewFlagSet("createwallet", flag.ExitOnError)
	listAddressesCmd := flag.NewFlagSet("listaddresses", flag.ExitOnError)

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

	case "createwallet":
		err := createWalletCmd.Parse(os.Args[2:])
		blockchain.ErrHandle(err)

	case "listaddresses":
		err := listAddressesCmd.Parse(os.Args[2:])
		blockchain.ErrHandle(err)

	default:
		cli.gracefullExit()
	}

	if getbalanceCmd.Parsed() {
		if *getBalanceAddress == "" {
			cli.Logger.Error("Address is required for getbalance command")
			cli.gracefullExit()
		}
		cli.getBalance(*getBalanceAddress)
	}
	if createblockchainCmd.Parsed() {
		if *createBlockChainAddress == "" {
			cli.Logger.Error("Address is required for createblockchain command")
			cli.gracefullExit()
		}
		cli.createBlockchain(*createBlockChainAddress)
	}

	if sendCmd.Parsed() {
		if *sendFrom == "" || *sendTo == "" || *sendAmount <= 0 {
			cli.Logger.Error("From, To and Amount are required for send command")
			cli.gracefullExit()
		}
		cli.send(*sendFrom, *sendTo, *sendAmount)
	}

	if createWalletCmd.Parsed() {
		cli.createWallet()
	}

	if listAddressesCmd.Parsed() {
		cli.listAddresses()
	}

	// Print chain
	if printChainCmd.Parsed() {
		cli.printChain()
	}
}
