//SPDX-License-Identifier: Unlicense
pragma solidity ^0.8.0;

import "../external_contracts/aave-v3-core/contracts/flashloan/base/FlashLoanSimpleReceiverBase.sol";
// import "../external_contracts/aave-v3-core/contracts/flashloan/interfaces/IFlashLoanSimpleReceiver.sol";
// import "../external_contracts/aave-v3-core/contracts/interfaces/IPool.sol";

contract Arbitrage is FlashLoanSimpleReceiverBase {
    function performArbitrage(address _pair, address token) public {
    }
}
