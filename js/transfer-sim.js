"use strict";

const mockReceiverBytecode =
  "0x608060405234801561000f575f5ffd5b50604080516370a0823160e01b81523060048201525f80359260203592903591906001600160a01b038516906370a0823190602401602060405180830381865afa15801561005f573d5f5f3e3d5ffd5b505050506040513d601f19601f82011682018060405250810190610083919061017d565b6040516323b872dd60e01b81526001600160a01b03858116600483015230602483015260448201859052919250908516906323b872dd906064016020604051808303815f875af11580156100d9573d5f5f3e3d5ffd5b505050506040513d601f19601f820116820180604052508101906100fd9190610194565b506040516370a0823160e01b81523060048201525f906001600160a01b038616906370a0823190602401602060405180830381865afa158015610142573d5f5f3e3d5ffd5b505050506040513d601f19601f82011682018060405250810190610166919061017d565b90505f61017383836101ba565b9050805f5260205ff35b5f6020828403121561018d575f5ffd5b5051919050565b5f602082840312156101a4575f5ffd5b815180151581146101b3575f5ffd5b9392505050565b818103818111156101d957634e487b7160e01b5f52601160045260245ffd5b9291505056fea26469706673582212204883c5e4b104bdfdfdb6fc0dafef3356b8ef02d5709f23c70b90c242773e33c064736f6c63430008210033";

function strip0x(hex) {
  if (typeof hex !== "string") return "";
  return hex.startsWith("0x") || hex.startsWith("0X") ? hex.slice(2) : hex;
}

function normalizeAddress(addr, name) {
  if (typeof addr !== "string") {
    throw new Error(`${name} must be a hex string`);
  }
  const hex = strip0x(addr);
  if (hex.length !== 40 || !/^[0-9a-fA-F]+$/.test(hex)) {
    throw new Error(`${name} must be a 20-byte hex address`);
  }
  return "0x" + hex.toLowerCase();
}

function pad32(hex) {
  return strip0x(hex).padStart(64, "0");
}

function toBigIntAmount(amount) {
  if (typeof amount === "bigint") return amount;
  if (typeof amount === "number") {
    if (!Number.isSafeInteger(amount)) {
      throw new Error("amount must be a safe integer or bigint");
    }
    return BigInt(amount);
  }
  if (typeof amount === "string") {
    return BigInt(amount);
  }
  if (amount && typeof amount.toString === "function") {
    return BigInt(amount.toString(10));
  }
  throw new Error("unsupported amount type");
}

function amountToHex(amount) {
  return toBigIntAmount(amount).toString(16);
}

function buildMockCallData(token, from, amount) {
  const tokenAddr = normalizeAddress(token, "token");
  const fromAddr = normalizeAddress(from, "from");
  const amtHex = amountToHex(amount);
  return "0x" + pad32(tokenAddr) + pad32(fromAddr) + pad32("0x" + amtHex);
}

function rpcRequest(provider, method, params) {
  if (!provider) {
    return Promise.reject(new Error("missing provider"));
  }

  if (typeof provider.request === "function") {
    return provider.request({ method, params });
  }

  return new Promise((resolve, reject) => {
    const payload = {
      jsonrpc: "2.0",
      id: Date.now(),
      method,
      params,
    };

    const callback = (err, res) => {
      if (err) return reject(err);
      if (res && res.error) return reject(res.error);
      return resolve(res ? res.result : undefined);
    };

    if (typeof provider.sendAsync === "function") {
      provider.sendAsync(payload, callback);
      return;
    }

    if (typeof provider.send === "function") {
      provider.send(payload, callback);
      return;
    }

    reject(new Error("provider does not support request/send"));
  });
}

async function transferSim(web3OrProvider, token, from, to, amount) {
  const provider =
    (web3OrProvider && (web3OrProvider.currentProvider || web3OrProvider.provider)) ||
    web3OrProvider;

  const toAddr = normalizeAddress(to, "to");
  const call = {
    to: toAddr,
    data: buildMockCallData(token, from, amount),
  };
  const overrides = {
    [toAddr]: { code: mockReceiverBytecode },
  };

  try {
    const result = await rpcRequest(provider, "eth_call", [call, "latest", overrides]);
    return { received: BigInt(result), error: null };
  } catch (err) {
    return { received: toBigIntAmount(amount), error: err };
  }
}

module.exports = {
  mockReceiverBytecode,
  buildMockCallData,
  transferSim,
};
