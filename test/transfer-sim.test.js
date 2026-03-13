"use strict";

const assert = require("node:assert/strict");
const test = require("node:test");
const Web3 = require("web3");

const { transferSim } = require("../js/transfer-sim");

test("returns the original amount object when the simulation fails", async () => {
  const amount = Web3.utils.toBN("123456789");
  const expectedError = new Error("rpc failed");
  const web3 = {
    eth: {
      abi: {
        encodeParameters(types, values) {
          assert.deepEqual(types, ["address", "address", "uint256"]);
          assert.equal(values[2], amount.toString());
          return "0xdeadbeef";
        }
      },
      extend({ methods }) {
        assert.equal(methods[0].call, "eth_call");
        this.callWithState = async () => {
          throw expectedError;
        };
      }
    }
  };

  const result = await transferSim(
    web3,
    "0x0000000000000000000000000000000000000001",
    "0x0000000000000000000000000000000000000002",
    "0x0000000000000000000000000000000000000003",
    amount
  );

  assert.equal(result.error, expectedError);
  assert.equal(result.received, amount);
});

test("returns zero without touching RPC for null or zero amount", async () => {
  for (const amount of [null, 0n]) {
    const result = await transferSim(
      {
        eth: {
          extend() {
            throw new Error("should not touch rpc");
          }
        }
      },
      "0x0000000000000000000000000000000000000001",
      "0x0000000000000000000000000000000000000002",
      "0x0000000000000000000000000000000000000003",
      amount
    );

    assert.equal(result.error, null);
    assert.equal(result.received, 0n);
  }
});
