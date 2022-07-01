package main

import (
	"fmt"

	"github.com/Xeway/mev-stuff/modules"
	"github.com/ethereum/go-ethereum/ethclient"
)

func main() {
	// Create a client instance to connect to our provider
	client, err := ethclient.Dial("https://eth-mainnet.alchemyapi.io/v2/YgladcqcW2iHbKZVTYd4S0VFcESU4UKw")

	if err != nil {
		fmt.Println(err)
	}

	amount := int64(1)

	graph := modules.GetAllRates(client, amount)

	fmt.Println(graph)

	// amountsOut := modules.GetAllUniswapAmountOut(client, amount)
	// bestOpportunity := modules.GetBestArbitrageOpportunity(client, amountsOut)
	// fmt.Println(bestOpportunity)
}
