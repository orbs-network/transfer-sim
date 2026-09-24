// Command runner exposes the Go implementation to the shared E2E scenarios.
package main

import (
	"encoding/json"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	transfersim "github.com/orbs-network/transfer-sim/go"
)

func main() {
	client, err := ethclient.Dial(os.Args[1])
	if err != nil {
		panic(err)
	}
	defer client.Close()
	amount, ok := new(big.Int).SetString(os.Args[5], 10)
	if !ok {
		panic("invalid amount")
	}
	received, callErr := transfersim.TransferSim(client,
		common.HexToAddress(os.Args[2]), common.HexToAddress(os.Args[3]), common.HexToAddress(os.Args[4]), amount)
	result := struct {
		Received string `json:"received"`
		Error    string `json:"error"`
	}{Received: received.String()}
	if callErr != nil {
		result.Error = callErr.Error()
	}
	// Mutating the returned integer must not mutate the caller's input.
	received.Add(received, big.NewInt(1))
	if amount.String() != os.Args[5] {
		panic("simulation aliased or mutated the input amount")
	}
	if err := json.NewEncoder(os.Stdout).Encode(result); err != nil {
		panic(err)
	}
}
