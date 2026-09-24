"use strict";

const assert = require("node:assert/strict");
const test = require("node:test");
const Web3 = require("web3");

const { transferSim } = require("../js/transfer-sim");

for (const [name, received] of [
  ["full transfer", (1n << 128n) + 123n],
  ["fee deducted", (1n << 128n) + 113n],
  ["entire amount deducted", 0n],
]) {
  test(name, async () => {
    const token = "0x0000000000000000000000000000000000000001";
    const from = "0x0000000000000000000000000000000000000002";
    const to = "0x0000000000000000000000000000000000000003";
    const amount = (1n << 128n) + 123n;
    const requests = [];
    const web3 = new Web3({
      send(payload, callback) {
        requests.push(payload);
        callback(null, { jsonrpc: "2.0", id: payload.id, result: "0x" + received.toString(16).padStart(64, "0") });
      },
    });
    assert.deepEqual(await transferSim(web3, token, from, to, amount), { received, error: null });
    assert.equal(requests.length, 1);
    const { method, params } = requests[0];
    assert.equal(method, "eth_call");
    assert.equal(params.length, 3);
    assert.deepEqual(params[0], {
      to,
      data: "0x" + token.slice(2).padStart(64, "0") + from.slice(2).padStart(64, "0") + amount.toString(16).padStart(64, "0"),
    });
    assert.equal(params[1], "latest");
    assert.deepEqual(Object.keys(params[2]), [to]);
    assert.deepEqual(Object.keys(params[2][to]), ["code"]);
    assert.match(params[2][to].code, /^0x(?:[0-9a-f]{2})+$/);
  });
}

test("returns the original amount and RPC revert", async () => {
  const web3 = new Web3({
    send(payload, callback) {
      callback(null, { jsonrpc: "2.0", id: payload.id, error: { code: 3, message: "execution reverted" } });
    },
  });
  const address = "0x0000000000000000000000000000000000000001";
  const result = await transferSim(web3, address, address, address, 123n);
  assert.equal(result.received, 123n);
  assert.match(result.error.message, /execution reverted/);
});

test("returns the original amount object when the simulation fails", async () => {
  const amount = Web3.utils.toBN("123456789");
  const expectedError = new Error("rpc failed");
  const web3 = new Web3({ send: (_, callback) => callback(expectedError) });

  const result = await transferSim(
    web3,
    "0x0000000000000000000000000000000000000001",
    "0x0000000000000000000000000000000000000002",
    "0x0000000000000000000000000000000000000003",
    amount
  );

  assert.match(result.error.message, /rpc failed/);
  assert.equal(result.received, amount);
});

test("returns zero without touching RPC for null or zero amount", async () => {
  for (const amount of [null, 0n]) {
    const result = await transferSim(
      null,
      "0x0000000000000000000000000000000000000001",
      "0x0000000000000000000000000000000000000002",
      "0x0000000000000000000000000000000000000003",
      amount
    );

    assert.equal(result.error, null);
    assert.equal(result.received, 0n);
  }
});
