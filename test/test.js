"use strict";

const Web3 = require("web3");
const { transferSim } = require("../js/transfer-sim");

async function main() {
  const rpcUrl = process.env.RPC_URL;
  if (!rpcUrl) throw new Error("missing RPC_URL");

  const web3 = new Web3(rpcUrl);

  const { TEST_TOKEN, TEST_FROM, TEST_TO, TEST_AMOUNT } = process.env;
  if (!TEST_TOKEN || !TEST_FROM || !TEST_TO || !TEST_AMOUNT) {
    throw new Error("missing TEST_* env vars");
  }
  const token = TEST_TOKEN;
  const from = TEST_FROM;
  const to = TEST_TO;
  const amount = BigInt(TEST_AMOUNT);

  const { received, error } = await transferSim(web3, token, from, to, amount);

  console.log("amount   :", amount.toString());
  console.log("received :", received.toString());
  console.log("error    :", error ? error.message || String(error) : "null");
}

main().catch((err) => {
  console.error(err);
  process.exitCode = 1;
});
