package transfersim

import (
	"context"
	"encoding/hex"
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

// TransferSim simulates a token transfer to detect fee-on-transfer behavior.
// It returns the ratio of actual transferred amount to expected amount as a fraction of 1e18.
// Returns 1e18 (1000000000000000000) if there's no fee on transfer.
// Returns less than 1e18 if there's a fee (e.g., 0.99e18 means 1% fee).
//
// Parameters:
//   - client: Ethereum RPC client
//   - token: ERC20 token contract address
//   - from: Address that has tokens and approval
//   - to: Destination address
//   - amount: Amount to transfer
//
// The function assumes that 'from' has already approved 'to' to spend tokens.
func TransferSim(client *ethclient.Client, token, from, to common.Address, amount *big.Int) (*big.Int, error) {
	ctx := context.Background()

	// Create mock contract bytecode that calls transferFrom and returns the balance difference
	mockContractCode := getMockReceiverBytecode()

	// Create state override for the 'to' address
	stateOverride := StateOverride{
		to: OverrideAccount{
			Code: (*hexutil.Bytes)(&mockContractCode),
		},
	}

	// Pack the call data: token, from, amount
	// The mock contract expects these as input parameters
	callData := make([]byte, 0, 96)
	callData = append(callData, common.LeftPadBytes(token.Bytes(), 32)...)
	callData = append(callData, common.LeftPadBytes(from.Bytes(), 32)...)
	callData = append(callData, common.LeftPadBytes(amount.Bytes(), 32)...)

	// Call with state override
	resultBytes, err := callContractWithStateOverride(client, ctx, &to, callData, nil, stateOverride)
	if err != nil {
		return nil, err
	}

	// Parse the result (actual amount transferred)
	actualTransferred := new(big.Int).SetBytes(resultBytes)

	// Calculate the ratio: actualTransferred / amount * 1e18
	// This gives us the percentage where 1e18 = 100% (no fee)
	ratio := new(big.Int).Mul(actualTransferred, big.NewInt(1e18))
	ratio.Div(ratio, amount)

	return ratio, nil
}

// getMockReceiverBytecode returns the compiled bytecode for MockReceiver contract
// This contract:
// 1. Receives token, from, and amount as calldata (96 bytes total)
// 2. Gets its balance before transfer
// 3. Calls transferFrom(from, this, amount) on the token
// 4. Gets its balance after transfer
// 5. Returns the difference (actual received amount)
func getMockReceiverBytecode() []byte {
	// This is the runtime bytecode compiled from MockReceiver.sol using solc 0.8.33
	// The fallback function reads token, from, amount from calldata
	// then executes the fee detection logic
	bytecodeHex := "608060405234801561000f575f5ffd5b50604080516370a0823160e01b81523060048201525f80359260203592903591906001600160a01b038516906370a0823190602401602060405180830381865afa15801561005f573d5f5f3e3d5ffd5b505050506040513d601f19601f82011682018060405250810190610083919061017d565b6040516323b872dd60e01b81526001600160a01b03858116600483015230602483015260448201859052919250908516906323b872dd906064016020604051808303815f875af11580156100d9573d5f5f3e3d5ffd5b505050506040513d601f19601f820116820180604052508101906100fd9190610194565b506040516370a0823160e01b81523060048201525f906001600160a01b038616906370a0823190602401602060405180830381865afa158015610142573d5f5f3e3d5ffd5b505050506040513d601f19601f82011682018060405250810190610166919061017d565b90505f61017383836101ba565b9050805f5260205ff35b5f6020828403121561018d575f5ffd5b5051919050565b5f602082840312156101a4575f5ffd5b815180151581146101b3575f5ffd5b9392505050565b818103818111156101d957634e487b7160e01b5f52601160045260245ffd5b9291505056fea26469706673582212204883c5e4b104bdfdfdb6fc0dafef3356b8ef02d5709f23c70b90c242773e33c064736f6c63430008210033"
	
	bytecode, _ := hex.DecodeString(bytecodeHex)
	return bytecode
}

// callContractWithStateOverride is a helper that calls a contract with state overrides
// This uses the eth_call RPC method with state override parameters
func callContractWithStateOverride(client *ethclient.Client, ctx context.Context, to *common.Address, data []byte, blockNumber *big.Int, overrides StateOverride) ([]byte, error) {
	// We need to use the raw RPC client to access state override functionality
	// The standard CallContract doesn't support state overrides
	
	type CallArgs struct {
		From                 *common.Address `json:"from,omitempty"`
		To                   *common.Address `json:"to,omitempty"`
		Gas                  *hexutil.Uint64 `json:"gas,omitempty"`
		GasPrice             *hexutil.Big    `json:"gasPrice,omitempty"`
		MaxFeePerGas         *hexutil.Big    `json:"maxFeePerGas,omitempty"`
		MaxPriorityFeePerGas *hexutil.Big    `json:"maxPriorityFeePerGas,omitempty"`
		Value                *hexutil.Big    `json:"value,omitempty"`
		Data                 *hexutil.Bytes  `json:"data,omitempty"`
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
