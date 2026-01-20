// Package transfersim simulates ERC20 transfers via state overrides to detect
// fee-on-transfer behavior without sending on-chain transactions.
package transfersim

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/ethclient"
)

// OverrideAccount specifies the fields to override in an account for state override
type OverrideAccount struct {
	Nonce     *hexutil.Uint64             `json:"nonce,omitempty"`
	Code      *hexutil.Bytes              `json:"code,omitempty"`
	Balance   *hexutil.Big                `json:"balance,omitempty"`
	State     map[common.Hash]common.Hash `json:"state,omitempty"`
	StateDiff map[common.Hash]common.Hash `json:"stateDiff,omitempty"`
}

// StateOverride is the collection of overridden accounts
type StateOverride map[common.Address]OverrideAccount

// TransferSim simulates a transferFrom and returns the amount received by `to`
// (balance delta). On RPC error it returns `amount` alongside the error.
// Requires `from` to have approved `to` to spend `amount`.
func TransferSim(client *ethclient.Client, token, from, to common.Address, amount *big.Int) (*big.Int, error) {
	ctx := context.Background()

	// Mock receiver bytecode calls transferFrom and returns its balance delta.
	mockContractCode := mockReceiverBytecode

	// Create state override for the 'to' address
	stateOverride := StateOverride{
		to: OverrideAccount{
			Code: (*hexutil.Bytes)(&mockContractCode),
		},
	}

	callData := buildMockCallData(token, from, amount)

	// Call with state override
	resultBytes, err := callWithStateOverride(client, ctx, &to, callData, nil, stateOverride)
	if err != nil {
		return new(big.Int).Set(amount), err
	}

	// Parse the result (actual amount received by destination)
	actualReceived := new(big.Int).SetBytes(resultBytes)

	return actualReceived, nil
}

func buildMockCallData(token, from common.Address, amount *big.Int) []byte {
	callData := make([]byte, 0, 96)
	callData = append(callData, common.LeftPadBytes(token.Bytes(), 32)...)
	callData = append(callData, common.LeftPadBytes(from.Bytes(), 32)...)
	callData = append(callData, common.LeftPadBytes(amount.Bytes(), 32)...)
	return callData
}

// Runtime bytecode for the mock receiver used in simulations.
var mockReceiverBytecode = hexutil.MustDecode("0x608060405234801561000f575f5ffd5b50604080516370a0823160e01b81523060048201525f80359260203592903591906001600160a01b038516906370a0823190602401602060405180830381865afa15801561005f573d5f5f3e3d5ffd5b505050506040513d601f19601f82011682018060405250810190610083919061017d565b6040516323b872dd60e01b81526001600160a01b03858116600483015230602483015260448201859052919250908516906323b872dd906064016020604051808303815f875af11580156100d9573d5f5f3e3d5ffd5b505050506040513d601f19601f820116820180604052508101906100fd9190610194565b506040516370a0823160e01b81523060048201525f906001600160a01b038616906370a0823190602401602060405180830381865afa158015610142573d5f5f3e3d5ffd5b505050506040513d601f19601f82011682018060405250810190610166919061017d565b90505f61017383836101ba565b9050805f5260205ff35b5f6020828403121561018d575f5ffd5b5051919050565b5f602082840312156101a4575f5ffd5b815180151581146101b3575f5ffd5b9392505050565b818103818111156101d957634e487b7160e01b5f52601160045260245ffd5b9291505056fea26469706673582212204883c5e4b104bdfdfdb6fc0dafef3356b8ef02d5709f23c70b90c242773e33c064736f6c63430008210033")

var callWithStateOverride = callContractWithStateOverride

// callContractWithStateOverride wraps eth_call with state overrides.
func callContractWithStateOverride(client *ethclient.Client, ctx context.Context, to *common.Address, data []byte, blockNumber *big.Int, overrides StateOverride) ([]byte, error) {
	// Use the raw RPC client since CallContract doesn't support overrides.
	type CallArgs struct {
		To   *common.Address `json:"to,omitempty"`
		Data *hexutil.Bytes  `json:"data,omitempty"`
	}

	callData := hexutil.Bytes(data)
	args := CallArgs{
		To:   to,
		Data: &callData,
	}

	var blockNumStr string
	if blockNumber == nil {
		blockNumStr = "latest"
	} else {
		blockNumStr = hexutil.EncodeBig(blockNumber)
	}

	var result hexutil.Bytes
	err := client.Client().CallContext(ctx, &result, "eth_call", args, blockNumStr, overrides)
	return result, err
}
