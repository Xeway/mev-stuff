package arb

import (
	"context"
	"log"
	"math/big"
	"sync"

	erc_20 "github.com/Xeway/mev-stuff/abi/go/erc_20"
	uniswap_router "github.com/Xeway/mev-stuff/abi/go/uniswap_router"
	"github.com/Xeway/mev-stuff/addresses"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

type ExchangeAndAmount struct {
	ExchangeAddr common.Address
	AmountOut    *big.Int
}

func GetAllUniswapAmountOut(client *ethclient.Client, stableAmount int) map[common.Address][]ExchangeAndAmount {
	// this map is structured like that : amountsOut[address of stablecoin like USDC] = an struct that contain the exchange + the amount out
	amountsOut := make(map[common.Address][]ExchangeAndAmount)

	executorWallet := common.HexToAddress(addresses.EXECUTOR_WALLET)
	options := &bind.CallOpts{Pending: true, From: executorWallet, BlockNumber: nil, Context: context.Background()}
	wethAddress := common.HexToAddress(addresses.WETH_ADDRESS)

	var wg1 sync.WaitGroup
	wg1.Add(len(addresses.ROUTER_ADDRESSES))

	for _, routerAddress := range addresses.ROUTER_ADDRESSES {
		go func(r string) {
			routerAddress := common.HexToAddress(r)

			instanceRouter, err := uniswap_router.NewUniswapRouterCaller(routerAddress, client)

			if err != nil {
				log.Fatal(err)
			}

			var wg2 sync.WaitGroup
			wg2.Add(len(addresses.STABLE_ADDRESSES))

			for _, stableAddress := range addresses.STABLE_ADDRESSES {
				go func(s string) {
					stableAddress := common.HexToAddress(s)

					stableInstance, err := erc_20.NewErc20Caller(stableAddress, client)
					if err != nil {
						log.Fatal(err)
					}

					stableDecimals, err := stableInstance.Decimals(options)
					if err != nil {
						log.Fatal(err)
					}

					amountIn := big.NewInt(0).Mul(big.NewInt(int64(stableAmount)), big.NewInt(0).Exp(big.NewInt(10), big.NewInt(int64(stableDecimals)), nil))

					amountOut, err := instanceRouter.GetAmountsOut(options, amountIn, []common.Address{stableAddress, wethAddress})
					if err != nil {
						log.Fatal(err)
					}

					amountsOut[stableAddress] = append(amountsOut[stableAddress], ExchangeAndAmount{
						ExchangeAddr: routerAddress,
						AmountOut:    amountOut[1],
					})

					wg2.Done()
				}(stableAddress)
			}
			wg2.Wait()

			wg1.Done()
		}(routerAddress)
	}

	wg1.Wait()

	return amountsOut
}

type Opportunity struct {
	Stable        common.Address
	BestAmount    *big.Int
	BestExchange  common.Address
	WorstAmount   *big.Int
	WorstExchange common.Address
}

func GetBestArbitrageOpportunity(client *ethclient.Client, amountsOut map[common.Address][]ExchangeAndAmount) Opportunity {
	var bestOpportunity Opportunity

	bestDelta := big.NewInt(0)

	for stableAddr, exchangeToAmount := range amountsOut { // we get an array that contains all amount and exchange address
		bestAmount := big.NewInt(0)
		var bestExchange common.Address

		worstAmount := big.NewInt(0)
		var worstExchange common.Address

		for i, exAndAm := range exchangeToAmount { // we get every struct ExchangeAndAmount
			if exAndAm.AmountOut.Cmp(bestAmount) == 1 {
				bestAmount = exAndAm.AmountOut
				bestExchange = exAndAm.ExchangeAddr
			}

			if exAndAm.AmountOut.Cmp(worstAmount) == -1 || i == 0 {
				worstAmount = exAndAm.AmountOut
				worstExchange = exAndAm.ExchangeAddr
			}
		}

		amountDelta := big.NewInt(0).Sub(bestAmount, worstAmount)

		if amountDelta.Cmp(bestDelta) == 1 {
			bestDelta = amountDelta

			bestOpportunity = Opportunity{
				Stable:        stableAddr,
				BestAmount:    bestAmount,
				BestExchange:  bestExchange,
				WorstAmount:   worstAmount,
				WorstExchange: worstExchange,
			}
		}

	}

	return bestOpportunity
}
