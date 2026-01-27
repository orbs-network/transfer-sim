"use strict";

const mockReceiverBytecode =
  "0x608060405234801561000f575f5ffd5b50604080516370a0823160e01b81523060048201525f80359260203592903591906001600160a01b038516906370a0823190602401602060405180830381865afa15801561005f573d5f5f3e3d5ffd5b505050506040513d601f19601f82011682018060405250810190610083919061017d565b6040516323b872dd60e01b81526001600160a01b03858116600483015230602483015260448201859052919250908516906323b872dd906064016020604051808303815f875af11580156100d9573d5f5f3e3d5ffd5b505050506040513d601f19601f820116820180604052508101906100fd9190610194565b506040516370a0823160e01b81523060048201525f906001600160a01b038616906370a0823190602401602060405180830381865afa158015610142573d5f5f3e3d5ffd5b505050506040513d601f19601f82011682018060405250810190610166919061017d565b90505f61017383836101ba565b9050805f5260205ff35b5f6020828403121561018d575f5ffd5b5051919050565b5f602082840312156101a4575f5ffd5b815180151581146101b3575f5ffd5b9392505050565b818103818111156101d957634e487b7160e01b5f52601160045260245ffd5b9291505056fea26469706673582212204883c5e4b104bdfdfdb6fc0dafef3356b8ef02d5709f23c70b90c242773e33c064736f6c63430008210033";

function ensureCallWithState(web3) {
  if (web3.eth.callWithState) return;
  web3.eth.extend({
    methods: [
      {
        name: "callWithState",
        call: "eth_call",
        params: 3,
      },
    ],
  });
}

async function transferSim(web3, token, from, to, amount, opts = {}) {
  try {
    ensureCallWithState(web3);

    const calldata = web3.eth.abi.encodeParameters(
        ["address", "address", "uint256"],
        [token, from, amount.toString()]
    );

    const result = await web3.eth.callWithState(
      {
        to,
        data: calldata
      },
      opts.blockNumber ? web3.eth.abi.encodeParameter("uint256", opts.blockNumber) : "latest",
      {
        [to]: { code: mockReceiverBytecode }
      }
    );

    const decoded = web3.eth.abi.decodeParameter("uint256", result);
    return { received: BigInt(decoded), error: null };
  } catch (err) {
    return { received: BigInt(amount), error: err };
  }
}

module.exports = { transferSim };
