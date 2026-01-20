# transfer-sim

A Go library for simulating off-chain token transfers to detect fee-on-transfer tokens using Ethereum state overrides.

## Overview

`transfer-sim` provides a single exported function `TransferSim` that uses geth's state override feature to detect whether an ERC20 token implements fee-on-transfer mechanics. This is useful for DeFi protocols and applications that need to handle tokens with transfer fees correctly.

## Features

- **State Override**: Uses Ethereum's `eth_call` with state overrides to simulate transfers without actual on-chain transactions
- **No Gas Cost**: Performs off-chain simulation, no gas required
- **Accurate Detection**: Measures actual transferred amount vs expected amount
- **Simple API**: Single function interface

## Installation

```bash
go get github.com/orbs-network/transfer-sim
```

## Usage

```go
package main

import (
    "fmt"
    "math/big"
    
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/ethclient"
    "github.com/orbs-network/transfer-sim"
)

func main() {
    // Connect to Ethereum node
    client, err := ethclient.Dial("https://eth-mainnet.alchemyapi.io/v2/YOUR-API-KEY")
    if err != nil {
        panic(err)
    }
    
    // Token and addresses
    token := common.HexToAddress("0x...") // ERC20 token address
    from := common.HexToAddress("0x...")  // Address with tokens and approval
    to := common.HexToAddress("0x...")    // Destination address
    amount := big.NewInt(1000000000000000000) // 1 token (18 decimals)
    
    // Run simulation
    ratio, err := transfersim.TransferSim(client, token, from, to, amount)
    if err != nil {
        panic(err)
    }
    
    // Interpret results
    oneE18 := big.NewInt(1e18)
    if ratio.Cmp(oneE18) == 0 {
        fmt.Println("No fee on transfer detected")
    } else {
        // Calculate fee percentage
        fee := new(big.Int).Sub(oneE18, ratio)
        feePercent := new(big.Int).Mul(fee, big.NewInt(100))
        feePercent.Div(feePercent, oneE18)
        fmt.Printf("Fee on transfer detected: %s%%\n", feePercent.String())
    }
}
```

## API

### TransferSim

```go
func TransferSim(
    client *ethclient.Client,
    token, from, to common.Address,
    amount *big.Int,
) (*big.Int, error)
```

Simulates a token transfer to detect fee-on-transfer behavior.

**Parameters:**
- `client`: Ethereum RPC client
- `token`: ERC20 token contract address
- `from`: Address that has tokens and approval (assumed to have already approved `to`)
- `to`: Destination address
- `amount`: Amount to transfer

**Returns:**
- Ratio of actual transferred amount to expected amount as a fraction of 1e18
- `1e18` (1000000000000000000) = no fee on transfer (100% received)
- Less than `1e18` = fee detected (e.g., `0.99e18` = 1% fee)

**Example Return Values:**
- `1000000000000000000` → No fee (100% of amount transferred)
- `990000000000000000` → 1% fee (99% of amount transferred)
- `950000000000000000` → 5% fee (95% of amount transferred)

## How It Works

1. The function creates bytecode for a mock contract
2. Uses Ethereum state override to temporarily replace the `to` address with this mock contract
3. The mock contract:
   - Checks its balance before the transfer
   - Calls `transferFrom(from, to, amount)` on the token
   - Checks its balance after the transfer
   - Returns the actual amount received (balance difference)
4. Calculates the ratio: `(actualReceived / amount) * 1e18`

## Requirements

- Go 1.19 or later
- Access to an Ethereum RPC endpoint
- The `from` address must have:
  - Sufficient token balance
  - Approval for the `to` address to spend tokens

## Testing

```bash
go test -v ./...
```

Note: Integration tests are skipped by default as they require an Ethereum node. To run them, set up a local node or testnet and modify the test files accordingly.

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! Please open an issue or submit a pull request.

