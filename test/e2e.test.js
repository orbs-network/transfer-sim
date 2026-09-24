"use strict";

const assert = require("node:assert/strict");
const { execFileSync, spawn } = require("node:child_process");
const { once } = require("node:events");
const { mkdtempSync, readFileSync, rmSync } = require("node:fs");
const { tmpdir } = require("node:os");
const path = require("node:path");
const test = require("node:test");
const Web3 = require("web3");
const { transferSim } = require("../js");

test("real RPC state overrides", { timeout: 120000 }, async (t) => {
  const temp = mkdtempSync(path.join(tmpdir(), "transfer-sim-"));
  t.after(() => rmSync(temp, { recursive: true, force: true }));
  execFileSync("forge", [
    "build", "--root", path.join(__dirname, "fixtures"), "--contracts", ".",
    "--out", path.join(temp, "out"), "--cache-path", path.join(temp, "cache"),
    "--use", "0.8.21", "--evm-version", "shanghai",
  ], { timeout: 60000, stdio: "pipe" });
  const artifact = JSON.parse(readFileSync(path.join(temp, "out/Token.sol/Token.json")));
  const goRunner = path.join(temp, "simulate");
  execFileSync("go", ["build", "-o", goRunner, "./test/runner"], { timeout: 60000 });

  const anvil = spawn("anvil", ["--host", "127.0.0.1", "--port", "0", "--hardfork", "shanghai", "--accounts", "2"], {
    stdio: ["ignore", "pipe", "ignore"],
  });
  t.after(async () => {
    if (anvil.pid && anvil.exitCode === null && anvil.signalCode === null) {
      const exited = once(anvil, "exit");
      anvil.kill();
      await exited;
    }
  });
  const rpcURL = await new Promise((resolve, reject) => {
    const timer = setTimeout(() => reject(new Error("Anvil startup timed out")), 10000);
    let output = "";
    anvil.on("error", (error) => { clearTimeout(timer); reject(error); });
    anvil.on("exit", () => { clearTimeout(timer); reject(new Error("Anvil exited")); });
    anvil.stdout.on("data", (chunk) => {
      output += chunk;
      const match = output.match(/Listening on (127\.0\.0\.1:\d+)/);
      if (match) { clearTimeout(timer); resolve(`http://${match[1]}`); }
    });
  });
  const web3 = new Web3(rpcURL);
  const [from, to] = await web3.eth.getAccounts();
  // Existing code makes the suite fail if the library omits the code override.
  await new Promise((resolve, reject) => web3.currentProvider.send({
    jsonrpc: "2.0", id: 1, method: "anvil_setCode", params: [to, "0x00"],
  }, (error, response) => error || response.error ? reject(error || response.error) : resolve()));
  assert.equal(await web3.eth.call({ to }), "0x", "receiver has no simulation logic without the override");

  const amount = (1n << 128n) + 123n;
  const scenarios = [
    { name: "full transfer", fee: 0, balance: amount * 2n, allowance: amount, received: amount },
    { name: "fee deducted", fee: 250, balance: amount * 2n, allowance: amount, received: amount - amount * 250n / 10000n },
    { name: "100% fee", fee: 10000, balance: amount * 2n, allowance: amount, received: 0n },
    { name: "insufficient balance", fee: 0, balance: amount - 1n, allowance: amount, error: "balance" },
    { name: "insufficient allowance", fee: 0, balance: amount * 2n, allowance: amount - 1n, error: "allowance" },
  ];
  const implementations = {
    js: async (token) => {
      const { received, error } = await transferSim(web3, token, from, to, amount);
      return { received: received.toString(), error: error ? error.message : "" };
    },
    go: (token) => JSON.parse(execFileSync(goRunner, [rpcURL, token, from, to, amount.toString()], { encoding: "utf8", timeout: 10000 })),
  };
  for (const scenario of scenarios) {
    await t.test(scenario.name, async (t) => {
      const token = await new web3.eth.Contract(artifact.abi).deploy({
        data: artifact.bytecode.object,
        arguments: [scenario.balance.toString(), to, scenario.fee],
      }).send({ from, gas: 1500000 });
      await token.methods.approve(to, scenario.allowance.toString()).send({ from, gas: 100000 });
      const state = () => Promise.all([
        token.methods.balanceOf(from).call(), token.methods.balanceOf(to).call(),
        token.methods.allowance(from, to).call(), web3.eth.getCode(to),
      ]);
      const before = await state();
      assert.deepEqual(before, [scenario.balance.toString(), "77", scenario.allowance.toString(), "0x00"]);
      for (const [name, simulate] of Object.entries(implementations)) {
        await t.test(name, async () => {
          const result = await simulate(token.options.address);
          assert.equal(result.received, (scenario.received ?? amount).toString());
          if (scenario.error) assert.match(result.error, new RegExp(scenario.error));
          else assert.equal(result.error, "");
          assert.deepEqual(await state(), before, "simulation must not persist state or code");
        });
      }
    });
  }
});
