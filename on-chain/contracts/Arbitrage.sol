//SPDX-License-Identifier: Unlicense
pragma solidity ^0.8.0;

import "../external_contracts/aave-v3-core/contracts/flashloan/base/FlashLoanSimpleReceiverBase.sol";
// import "../external_contracts/aave-v3-core/contracts/flashloan/interfaces/IFlashLoanSimpleReceiver.sol";
// import "../external_contracts/aave-v3-core/contracts/interfaces/IPool.sol";

import "@openzeppelin/contracts/token/ERC20/IERC20.sol";

contract Arbitrage is FlashLoanSimpleReceiverBase {
    constructor(address _providerAddress)
    FlashLoanSimpleReceiverBase(_providerAddress)
    public {}

    function performArbitrage(address _pair, address _token, uint _amount) public {
        POOL.flashLoanSimple(address(this), _token, _amount, bytes(""), 0);
    }

    function executeOperation(
        address asset,
        uint256 amount,
        uint256 premium,
        address initiator,
        bytes calldata params
    ) external override returns (bool) {
        // arbitrage code will be here

        uint amountOwing = amount + premium;
        IERC20(asset).approve(address(POOL), amountOwing);

        return true;
    }
}
