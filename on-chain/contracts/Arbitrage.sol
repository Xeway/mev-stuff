//SPDX-License-Identifier: Unlicense
pragma solidity ^0.8.0;

import "../external_contracts/aave-v3-core/contracts/flashloan/base/FlashLoanSimpleReceiverBase.sol";
// import "../external_contracts/aave-v3-core/contracts/flashloan/interfaces/IFlashLoanSimpleReceiver.sol";
// import "../external_contracts/aave-v3-core/contracts/interfaces/IPool.sol";

import "./interfaces/IUniswapV2Router02.sol";

import "@openzeppelin/contracts/token/ERC20/IERC20.sol";

contract Arbitrage is FlashLoanSimpleReceiverBase {
    constructor(address _providerAddress)
    FlashLoanSimpleReceiverBase(_providerAddress)
    public {}

    function performArbitrage(
        address _tokenToBorrow,
        uint _amountToBorrow,
        address _tokenToReceive,
        uint _amountToReceiveMin,
        address _routerAddressToBorrow,
        address _routerAddressToReceive
    ) public {
        bytes parameters = abi.encode(
            _tokenToBorrow,
            _amountToBorrow,
            _tokenToReceive,
            _amountToReceiveMin,
            _routerAddressToBorrow,
            _routerAddressToReceive
        );

        POOL.flashLoanSimple(address(this), _tokenToBorrow, _amountToBorrow, parameters, 0);
    }

    function executeOperation(
        address asset,
        uint256 amount,
        uint256 premium,
        address initiator,
        bytes calldata params
    ) external override returns (bool) {
        // arbitrage code will be here
        (
            _tokenToBorrow,
            _amountToBorrow,
            _tokenToReceive,
            _amountToReceiveMin,
            _routerAddressToBorrow,
            _routerAddressToReceive
        ) = abi.decode(params, (address, uint, address, uint, address, address));
        
        // first swap
        _swap(_tokenToBorrow, _amountToBorrow, _tokenToReceive, _amountToReceiveMin, _routerAddressToBorrow);
        // at this point, we've received at least _amountToReceive token

        // second swap
        _swap(_tokenToReceive, _amountToReceiveMin, _tokenToBorrow, _amountToBorrow, _amountToBorrow, _routerAddressToReceive);

        uint amountOwing = amount + premium;
        IERC20(asset).approve(address(POOL), amountOwing);

        return true;
    }

    function _swap(address _tokenIn, uint _amountIn, address _tokenOut, uint _amountOutMin, address _routerAddress) internal {
        IERC20(_tokenIn).approve(_routerAddress, _amountIn);

        address[] memory path = new address[](2);
        path[0] = _tokenIn;
        path[1] = _tokenOut;

        IUniswapV2Router02(_routerAddress)
        .swapExactTokensForTokensSupportingFeeOnTransferTokens(
            _amountIn,
            _amountOutMin,
            path,
            address(this),
            block.timestamp
        );
    }
}
