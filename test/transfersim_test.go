package transfersim_test

import (
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/ethclient"
	transfersim "github.com/orbs-network/transfer-sim/go"
)

func TestTransferSimSuccess(t *testing.T) {
	token := common.HexToAddress("0x1")
	from := common.HexToAddress("0x2")
	to := common.HexToAddress("0x3")
	amount := new(big.Int).Add(new(big.Int).Lsh(big.NewInt(1), 128), big.NewInt(123))
	for _, tc := range []struct {
		name     string
		received *big.Int
	}{
		{"full transfer", new(big.Int).Set(amount)},
		{"fee deducted", new(big.Int).Sub(amount, big.NewInt(10))},
		{"entire amount deducted", big.NewInt(0)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			requests := make(chan []json.RawMessage, 1)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var req struct {
					ID     json.RawMessage   `json:"id"`
					Method string            `json:"method"`
					Params []json.RawMessage `json:"params"`
				}
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					t.Error(err)
					return
				}
				if req.Method != "eth_call" {
					t.Errorf("unexpected method: %s", req.Method)
				}
				requests <- req.Params
				w.Header().Set("Content-Type", "application/json")
				if err := json.NewEncoder(w).Encode(map[string]any{
					"jsonrpc": "2.0", "id": req.ID, "result": fmt.Sprintf("0x%064x", tc.received),
				}); err != nil {
					t.Error(err)
				}
			}))
			defer server.Close()
			client, err := ethclient.Dial(server.URL)
			if err != nil {
				t.Fatal(err)
			}
			defer client.Close()
			received, err := transfersim.TransferSim(client, token, from, to, amount)
			if err != nil {
				t.Fatal(err)
			}
			if received.Cmp(tc.received) != 0 {
				t.Fatalf("received %v, want %v", received, tc.received)
			}
			params := <-requests
			if len(params) != 3 {
				t.Fatalf("got %d params", len(params))
			}
			var call struct {
				To   common.Address
				Data hexutil.Bytes
			}
			if err := json.Unmarshal(params[0], &call); err != nil {
				t.Fatal(err)
			}
			wantData := fmt.Sprintf("%064x%064x%064x", token.Bytes(), from.Bytes(), amount)
			if call.To != to || fmt.Sprintf("%x", []byte(call.Data)) != wantData {
				t.Fatalf("unexpected call: %+v", call)
			}
			if string(params[1]) != `"latest"` {
				t.Fatalf("unexpected block: %s", params[1])
			}
			var overrides map[common.Address]map[string]hexutil.Bytes
			if err := json.Unmarshal(params[2], &overrides); err != nil {
				t.Fatal(err)
			}
			if len(overrides) != 1 || len(overrides[to]) != 1 || len(overrides[to]["code"]) == 0 {
				t.Fatalf("unexpected overrides: %v", overrides)
			}
		})
	}
}

func TestTransferSimReturnsOriginalAmountOnRPCError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		var req struct {
			ID     any    `json:"id"`
			Method string `json:"method"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("decode request: %v", err)
			return
		}
		if req.Method != "eth_call" {
			t.Errorf("expected eth_call, got %s", req.Method)
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
			t.Errorf("encode response: %v", err)
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
