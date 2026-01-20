package transfersim

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

// TestMockReceiverBytecode ensures the embedded bytecode is present.
func TestMockReceiverBytecode(t *testing.T) {
	if len(mockReceiverBytecode) == 0 {
		t.Fatal("mock receiver bytecode is empty")
	}
	if len(mockReceiverBytecode) < 100 {
		t.Fatalf("mock receiver bytecode too small: %d bytes", len(mockReceiverBytecode))
	}
}

func TestBuildMockCallData(t *testing.T) {
	token := common.HexToAddress("0x1111111111111111111111111111111111111111")
	from := common.HexToAddress("0x2222222222222222222222222222222222222222")
	amount := big.NewInt(0x1234)

	got := buildMockCallData(token, from, amount)
	if len(got) != 96 {
		t.Fatalf("call data length = %d, want 96", len(got))
	}

	expectedHex := "" +
		"0000000000000000000000001111111111111111111111111111111111111111" +
		"0000000000000000000000002222222222222222222222222222222222222222" +
		"0000000000000000000000000000000000000000000000000000000000001234"
	expected, err := hex.DecodeString(expectedHex)
	if err != nil {
		t.Fatalf("decode expected hex: %v", err)
	}

	if !bytes.Equal(got, expected) {
		t.Fatalf("call data mismatch\n got: %x\nwant: %x", got, expected)
	}
}

func TestTransferSimErrorReturnsAmount(t *testing.T) {
	orig := callWithStateOverride
	t.Cleanup(func() { callWithStateOverride = orig })
	callWithStateOverride = func(_ *ethclient.Client, _ context.Context, _ *common.Address, _ []byte, _ *big.Int, _ StateOverride) ([]byte, error) {
		return nil, errors.New("boom")
	}

	token := common.HexToAddress("0x1111111111111111111111111111111111111111")
	from := common.HexToAddress("0x2222222222222222222222222222222222222222")
	to := common.HexToAddress("0x3333333333333333333333333333333333333333")
	amount := big.NewInt(123)

	got, err := TransferSim(nil, token, from, to, amount)
	if err == nil {
		t.Fatal("expected error")
	}
	if got == nil || got.Cmp(amount) != 0 {
		t.Fatalf("got %v, want %s", got, amount.String())
	}
}
