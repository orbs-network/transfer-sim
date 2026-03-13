package main

import (
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	transfersim "github.com/orbs-network/transfer-sim/go"
)

func TestTransferSimReturnsOriginalAmountOnRPCError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		var req struct {
			ID     any    `json:"id"`
			Method string `json:"method"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if req.Method != "eth_call" {
			t.Fatalf("expected eth_call, got %s", req.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]any{
			"jsonrpc": "2.0",
			"id":      req.ID,
			"error": map[string]any{
				"code":    3,
				"message": "execution reverted",
			},
		}); err != nil {
			t.Fatalf("encode response: %v", err)
		}
	}))
	defer server.Close()

	client, err := ethclient.Dial(server.URL)
	if err != nil {
		t.Fatalf("dial test rpc: %v", err)
	}
	defer client.Close()

	amount := big.NewInt(123456789)
	received, err := transfersim.TransferSim(
		client,
		common.Address{},
		common.Address{},
		common.Address{},
		amount,
	)
	if err == nil {
		t.Fatal("expected rpc error")
	}
	if !strings.Contains(err.Error(), "execution reverted") {
		t.Fatalf("expected revert error, got %v", err)
	}
	if received == nil || received.Cmp(amount) != 0 {
		t.Fatalf("expected %v, got %v", amount, received)
	}

	received.Add(received, big.NewInt(1))
	if amount.String() != "123456789" {
		t.Fatalf("expected original amount to remain unchanged, got %v", amount)
	}
}

func TestTransferSimReturnsZeroForNilOrZeroAmount(t *testing.T) {
	cases := []struct {
		name   string
		amount *big.Int
	}{
		{name: "nil", amount: nil},
		{name: "zero", amount: big.NewInt(0)},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			received, err := transfersim.TransferSim(
				nil,
				common.Address{},
				common.Address{},
				common.Address{},
				tc.amount,
			)
			if err != nil {
				t.Fatalf("expected nil error, got %v", err)
			}
			if received == nil || received.Sign() != 0 {
				t.Fatalf("expected zero, got %v", received)
			}
		})
	}
}

func TestTransferSimReturnsOriginalAmountOnPanic(t *testing.T) {
	amount := big.NewInt(42)
	received, err := transfersim.TransferSim(
		nil,
		common.Address{},
		common.Address{},
		common.Address{},
		amount,
	)
	if err == nil {
		t.Fatal("expected recovered panic error")
	}
	if received == nil || received.Cmp(amount) != 0 {
		t.Fatalf("expected %v, got %v", amount, received)
	}
	if !strings.Contains(err.Error(), "transfer simulation failed") {
		t.Fatalf("expected recovered panic message, got %v", err)
	}
}
