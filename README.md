# transfer-sim

🧪 Simulate ERC20 transfers via state overrides to detect fee-on-transfer tokens.

## ⚡ Install

```bash
go get github.com/orbs-network/transfer-sim
```

## ✅ Usage

```go
client, _ := ethclient.Dial("https://eth-mainnet.alchemyapi.io/v2/YOUR-API-KEY")
token := common.HexToAddress("0x...")
from := common.HexToAddress("0x...")
to := common.HexToAddress("0x...")
amount := big.NewInt(1000000000000000000) // 1 token (18 decimals)

received, err := transfersim.TransferSim(client, token, from, to, amount)
if err != nil {
    // On error, received == amount. Treat as unknown.
    fmt.Printf("Simulation error: %v\n", err)
    return
}

if received.Cmp(amount) == 0 {
    fmt.Println("No fee on transfer detected")
} else {
    fee := new(big.Int).Sub(amount, received)
    fmt.Printf("Fee on transfer detected: %s tokens\n", fee.String())
}
```

## 🧰 API

```go
func TransferSim(
    client *ethclient.Client,
    token, from, to common.Address,
    amount *big.Int,
) (*big.Int, error)
```

- Returns the actual amount received by `to` (balance delta).
- On RPC error, returns `amount` alongside the error.
- Requires `from` to approve `to` to spend `amount`.
- Uses `eth_call` with state overrides (no on-chain tx).

## 🧪 Tests

```bash
go test ./...
```
