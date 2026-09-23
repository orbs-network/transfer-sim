"use strict";

const Web3 = require("web3");
const { transferSim } = require("../js/transfer-sim");

async function main() {
  const web3 = new Web3(process.env.ETH_RPC_URL);
  const from = process.env.ETH_FROM;

  const token = "0x8ac76a51cc950d9822d68b83fe1ad97b32cd580d";
  const to = "0x00002a9C4D9497df5Bd31768eC5d30eEf5405000";
  const amount = 123456789n;

  const { received, error } = await transferSim(web3, token, from, to, amount);

  console.log("amount   :", amount.toString());
  console.log("received :", received.toString());
  console.log("error    :", error ? error.message || String(error) : "null");
}

main().catch((err) => {
  console.error(err);
  process.exitCode = 1;
});
