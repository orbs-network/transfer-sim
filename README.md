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

## Node.js Usage

The Go library is the canonical implementation. A Node.js translation is provided in `js/transfer-sim.js` for convenience.

```js
const Web3 = require("web3");
const { transferSim } = require("./js/transfer-sim");

const web3 = new Web3(process.env.RPC_URL);

const token = "0x...";
const from = "0x...";
const to = "0x...";
const amount = 10n ** 18n;

(async () => {
  const { received, error } = await transferSim(web3, token, from, to, amount);
  if (error) {
    console.warn("Simulation error:", error.message || error);
  }
  console.log("received:", received.toString());
})();
```

- Requires `from` to approve `to` to spend `amount` on-chain.
- Requires an RPC that supports `eth_call` with state overrides.

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
