package transfersim_test

import (
	"math/big"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	transfersim "github.com/orbs-network/transfer-sim/go"
)

func TestTransferSimReturnsZeroWithoutRPC(t *testing.T) {
	for _, amount := range []*big.Int{nil, big.NewInt(0)} {
		received, err := transfersim.TransferSim(nil, common.Address{}, common.Address{}, common.Address{}, amount)
		if err != nil || received == nil || received.Sign() != 0 {
			t.Fatalf("expected zero and no error, got %v, %v", received, err)
		}
	}
}

func TestTransferSimRecoversPanic(t *testing.T) {
	amount := big.NewInt(42)
	received, err := transfersim.TransferSim(nil, common.Address{}, common.Address{}, common.Address{}, amount)
	if err == nil || !strings.Contains(err.Error(), "transfer simulation failed") {
		t.Fatalf("expected recovered panic, got %v", err)
	}
	if received == nil || received.Cmp(amount) != 0 {
		t.Fatalf("expected %v, got %v", amount, received)
	}
}
