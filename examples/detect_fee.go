package main

import (
	"fmt"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/orbs-network/transfer-sim"
)

func main() {
	// Get RPC URL from environment or use default
	rpcURL := os.Getenv("ETH_RPC_URL")
	if rpcURL == "" {
		rpcURL = "http://localhost:8545" // Default to local node
	}

	// Connect to Ethereum node
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		fmt.Printf("Failed to connect to Ethereum node: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()

	// Example addresses (replace with actual addresses)
	// For testing, you can use addresses from a local fork or testnet
	tokenAddr := os.Getenv("TOKEN_ADDRESS")
	fromAddr := os.Getenv("FROM_ADDRESS")
	toAddr := os.Getenv("TO_ADDRESS")

	if tokenAddr == "" || fromAddr == "" || toAddr == "" {
		fmt.Println("Please set environment variables:")
		fmt.Println("  TOKEN_ADDRESS - ERC20 token contract address")
		fmt.Println("  FROM_ADDRESS - Address with tokens and approval")
		fmt.Println("  TO_ADDRESS - Destination address")
		fmt.Println("  ETH_RPC_URL - Ethereum RPC endpoint (optional, defaults to localhost:8545)")
		os.Exit(1)
	}

	token := common.HexToAddress(tokenAddr)
	from := common.HexToAddress(fromAddr)
	to := common.HexToAddress(toAddr)

	// Transfer amount: 1 token with 18 decimals
	amount := new(big.Int)
	amount.SetString("1000000000000000000", 10) // 1e18

	fmt.Printf("Simulating transfer of %s tokens\n", amount.String())
	fmt.Printf("Token: %s\n", token.Hex())
	fmt.Printf("From: %s\n", from.Hex())
	fmt.Printf("To: %s\n", to.Hex())
	fmt.Println()

	// Run the simulation
	ratio, err := transfersim.TransferSim(client, token, from, to, amount)
	if err != nil {
		fmt.Printf("Simulation failed: %v\n", err)
		os.Exit(1)
	}

	// Interpret the results
	oneE18 := new(big.Int)
	oneE18.SetString("1000000000000000000", 10) // 1e18

	fmt.Printf("Result ratio: %s (out of %s)\n", ratio.String(), oneE18.String())
	fmt.Println()

	if ratio.Cmp(oneE18) == 0 {
		fmt.Println("✓ No fee on transfer detected")
		fmt.Println("  100% of tokens are transferred")
	} else if ratio.Cmp(oneE18) < 0 {
		// Calculate fee percentage
		fee := new(big.Int).Sub(oneE18, ratio)
		
		// Convert to percentage with 2 decimal places
		feePercent := new(big.Int).Mul(fee, big.NewInt(10000))
		feePercent.Div(feePercent, oneE18)
		
		intPart := new(big.Int).Div(feePercent, big.NewInt(100))
		decPart := new(big.Int).Mod(feePercent, big.NewInt(100))
		
		fmt.Printf("✗ Fee on transfer detected: %s.%02s%%\n", intPart.String(), decPart.String())
		
		// Calculate actual received amount
		actualReceived := new(big.Int).Mul(amount, ratio)
		actualReceived.Div(actualReceived, oneE18)
		
		fmt.Printf("  Actual amount received: %s (out of %s)\n", actualReceived.String(), amount.String())
	} else {
		// This shouldn't happen, but handle it just in case
		fmt.Println("⚠ Unexpected result: ratio > 1e18")
		fmt.Println("  This might indicate a rebasing token or other unusual behavior")
	}
}
