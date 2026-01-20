package transfersim

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

func TestTransferSimIntegration(t *testing.T) {
	const (
		rpcURL    = "https://ethereum-rpc.publicnode.com"
		tokenAddr = "0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2" // WETH
		fromAddr  = "0x0000000000000000000000000000000000000001"
		toAddr    = "0x0000000000000000000000000000000000000002"
	)
	amount := big.NewInt(0) // zero amount avoids allowance/balance requirements

	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer client.Close()

	received, err := TransferSim(
		client,
		common.HexToAddress(tokenAddr),
		common.HexToAddress(fromAddr),
		common.HexToAddress(toAddr),
		amount,
	)
	if err != nil {
		t.Fatalf("TransferSim: %v", err)
	}
	if received == nil {
		t.Fatal("received is nil")
	}
	if received.Sign() < 0 {
		t.Fatalf("received negative: %s", received.String())
	}
	t.Logf("received=%s amount=%s", received.String(), amount.String())
}
