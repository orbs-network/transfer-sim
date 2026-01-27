package main

import (
	"fmt"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	transfersim "github.com/orbs-network/transfer-sim/go"
)

func main() {
	rpcURL := os.Getenv("RPC_URL")
	if rpcURL == "" {
		fmt.Println("missing RPC_URL")
		os.Exit(1)
	}

	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		fmt.Println("dial error:", err)
		os.Exit(1)
	}
	defer client.Close()

	tokenStr := os.Getenv("TEST_TOKEN")
	fromStr := os.Getenv("TEST_FROM")
	toStr := os.Getenv("TEST_TO")
	amountStr := os.Getenv("TEST_AMOUNT")
	if tokenStr == "" || fromStr == "" || toStr == "" || amountStr == "" {
		fmt.Println("missing TEST_* env vars")
		os.Exit(1)
	}

	token := common.HexToAddress(tokenStr)
	from := common.HexToAddress(fromStr)
	to := common.HexToAddress(toStr)
	amount, ok := new(big.Int).SetString(amountStr, 10)
	if !ok {
		fmt.Println("invalid TEST_AMOUNT:", amountStr)
		os.Exit(1)
	}

	received, callErr := transfersim.TransferSim(client, token, from, to, amount)

	fmt.Println("amount   :", amount.String())
	if received != nil {
		fmt.Println("received :", received.String())
	} else {
		fmt.Println("received :", "<nil>")
	}
	if callErr != nil {
		fmt.Println("error    :", callErr)
	} else {
		fmt.Println("error    :", "null")
	}
}
