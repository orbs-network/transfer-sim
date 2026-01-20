// Package transfersim provides utilities for detecting fee-on-transfer tokens
// using Ethereum state overrides to simulate transfers off-chain.
//
// The main function TransferSim uses geth's state override feature to
// temporarily replace a destination address with a mock contract that
// measures the actual amount of tokens received during a transferFrom call.
// This allows detection of tokens that deduct fees during transfers without
// requiring actual on-chain transactions.
//
// Example usage:
//
//	client, _ := ethclient.Dial("https://eth-mainnet.alchemyapi.io/v2/YOUR-API-KEY")
//	token := common.HexToAddress("0x...")
//	from := common.HexToAddress("0x...")
//	to := common.HexToAddress("0x...")
//	amount := big.NewInt(1000000000000000000) // 1 token
//
//	ratio, err := transfersim.TransferSim(client, token, from, to, amount)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// ratio == 1e18 means no fee, < 1e18 means fee detected
//	if ratio.Cmp(big.NewInt(1e18)) < 0 {
//	    fmt.Println("Fee on transfer detected")
//	}
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
//
// The bytecode was compiled from the following Solidity code:
//
//	// SPDX-License-Identifier: MIT
//	pragma solidity ^0.8.0;
//
//	interface IERC20 {
//	    function balanceOf(address account) external view returns (uint256);
//	    function transferFrom(address from, address to, uint256 amount) external returns (bool);
//	}
//
//	contract MockReceiver {
//	    fallback() external {
//	        address token;
//	        address from;
//	        uint256 amount;
//
//	        assembly {
//	            token := calldataload(0)
//	            from := calldataload(32)
//	            amount := calldataload(64)
//	        }
//
//	        uint256 balanceBefore = IERC20(token).balanceOf(address(this));
//	        IERC20(token).transferFrom(from, address(this), amount);
//	        uint256 balanceAfter = IERC20(token).balanceOf(address(this));
//	        uint256 actualReceived = balanceAfter - balanceBefore;
//
//	        assembly {
//	            mstore(0, actualReceived)
//	            return(0, 32)
//	        }
//	    }
//	}
//
// Compiled with: solc 0.8.33 with optimization enabled
// Command: solcjs --bin --optimize MockReceiver.sol
func getMockReceiverBytecode() []byte {
	// Runtime bytecode (deployed code)
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
