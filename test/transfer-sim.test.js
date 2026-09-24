"use strict";

const assert = require("node:assert/strict");
const test = require("node:test");
const Web3 = require("web3");
const { transferSim } = require("../js");

const address = "0x0000000000000000000000000000000000000001";

test("preserves the original amount object on transport failure", async () => {
  const amount = Web3.utils.toBN("123456789");
  const web3 = new Web3({ send: (_, callback) => callback(new Error("rpc failed")) });
  const result = await transferSim(web3, address, address, address, amount);
  assert.match(result.error.message, /rpc failed/);
  assert.equal(result.received, amount);
});

test("returns zero without touching RPC for null or zero amount", async () => {
  for (const amount of [null, 0n]) {
    assert.deepEqual(await transferSim(null, address, address, address, amount), {
      received: 0n, error: null,
    });
  }
});
