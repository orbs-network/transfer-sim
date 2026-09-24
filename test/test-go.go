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
	client, err := ethclient.Dial(os.Getenv("ETH_RPC_URL"))
	if err != nil {
		fmt.Println("dial error:", err)
		os.Exit(1)
	}
	defer client.Close()

	const tokenStr = "0x8ac76a51cc950d9822d68b83fe1ad97b32cd580d"
	const toStr = "0x00002a9C4D9497df5Bd31768eC5d30eEf5405000"
	const amountValue = 123456789
	token := common.HexToAddress(tokenStr)
	from := common.HexToAddress(os.Getenv("ETH_FROM"))
	to := common.HexToAddress(toStr)
	amount := big.NewInt(amountValue)

	received, callErr := transfersim.TransferSim(client, token, from, to, amount)

	fmt.Println("amount   :", amount.String())
	if received != nil {
		fmt.Println("received :", received.String())
	} else {
		fmt.Println("received :", "<nil>")
	}
	if callErr != nil {
		fmt.Println("error    :", callErr)
		os.Exit(1)
	} else {
		fmt.Println("error    :", "null")
	}
}
