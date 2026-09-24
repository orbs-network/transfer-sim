// SPDX-License-Identifier: MIT
pragma solidity 0.8.21;

// Minimal token fixture: real balances, approvals, and a configurable transfer fee.
contract Token {
    mapping(address => uint256) public balanceOf;
    mapping(address => mapping(address => uint256)) public allowance;
    uint256 public immutable feeBps;

    constructor(uint256 supply, address receiver, uint256 fee) {
        balanceOf[msg.sender] = supply;
        balanceOf[receiver] = 77;
        feeBps = fee;
    }

    function approve(address spender, uint256 amount) external returns (bool) {
        allowance[msg.sender][spender] = amount;
        return true;
    }

    function transferFrom(address from, address to, uint256 amount) external returns (bool) {
        require(allowance[from][msg.sender] >= amount, "allowance");
        require(balanceOf[from] >= amount, "balance");
        allowance[from][msg.sender] -= amount;
        balanceOf[from] -= amount;
        balanceOf[to] += amount - amount * feeBps / 10000;
        return true;
    }
}
