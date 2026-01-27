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

type overrideAccount struct {
	Code hexutil.Bytes `json:"code,omitempty"`
}

type stateOverride map[common.Address]overrideAccount

// TransferSim simulates a transferFrom and returns the amount received by `to`
// (balance delta). On RPC error it returns `amount` alongside the error.
// Requires `from` to have approved `to` to spend `amount`.
func TransferSim(client *ethclient.Client, token, from, to common.Address, amount *big.Int) (*big.Int, error) {
	callData := buildMockCallData(token, from, amount)
	overrides := stateOverride{
		to: {Code: mockReceiverBytecode},
	}

	resultBytes, err := callWithStateOverride(client, to, callData, overrides)
	if err != nil {
		return new(big.Int).Set(amount), err
	}

	return new(big.Int).SetBytes(resultBytes), nil
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

var callWithStateOverride = func(client *ethclient.Client, to common.Address, data []byte, overrides stateOverride) ([]byte, error) {
	type callArgs struct {
		To   common.Address `json:"to"`
		Data hexutil.Bytes  `json:"data"`
	}

	var result hexutil.Bytes
	err := client.Client().CallContext(
		context.Background(),
		&result,
		"eth_call",
		callArgs{To: to, Data: data},
		"latest",
		overrides,
	)
	return result, err
}
