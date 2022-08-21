package main

import (
	"context"
	"fmt"
	"time"

	"github.com/Xeway/mev-stuff/addresses"
	"github.com/Xeway/mev-stuff/modules"
	"github.com/Xeway/mev-stuff/query"
	"github.com/ethereum/go-ethereum/ethclient"
)

func QueryBiggestPairs() {
	pairs := query.GetPairsWithMostReserves()

	for i := 0; i < len(pairs); i++ {
		if modules.CheckIfNotPresentInArray(pairs[i].Token0.Id, addresses.TOKEN_ADDRESSES) {
			addresses.TOKEN_ADDRESSES = append(addresses.TOKEN_ADDRESSES, pairs[i].Token0.Id)
		}

		if modules.CheckIfNotPresentInArray(pairs[i].Token1.Id, addresses.TOKEN_ADDRESSES) {
			addresses.TOKEN_ADDRESSES = append(addresses.TOKEN_ADDRESSES, pairs[i].Token1.Id)
		}
	}
}

func EvaluateArb(client *ethclient.Client) {
	amount := int64(1)

	graph := modules.GetAllRates(client, amount)
	bestPath, err := modules.FindBestPath(graph)
	// if no arbitrage opportunity, try again
	if err != nil {
		fmt.Println(err)
		EvaluateArb(client)
	}

	fmt.Println(graph, "\n\n\n", bestPath)

	// amountsOut := modules.GetAllUniswapAmountOut(client, amount)
	// bestOpportunity := modules.GetBestArbitrageOpportunity(client, amountsOut)
	// fmt.Println(bestOpportunity)
}

func main() {
	start := time.Now()

	// Create a client instance to connect to our provider
	client, err := ethclient.Dial("wss://mainnet.infura.io/ws/v3/7f241e16d45245599aedb55e901250c2")
	if err != nil {
		fmt.Println(err)
	}

	QueryBiggestPairs()

	currentHeader, _ := client.HeaderByNumber(context.Background(), nil)
	currentBlock := currentHeader.Number.Int64()

	EvaluateArb(client)

	latestHeader, _ := client.HeaderByNumber(context.Background(), nil)
	latestBlock := latestHeader.Number.Int64()
	// if the block at the beginning of the computation is different from the block after the computation
	// it means that data changed and therefore arbitrage opportunity can be biased
	// so in this case we go the computation again
	if currentBlock != latestBlock {
		EvaluateArb(client)
	}

	fmt.Println(time.Since(start))
}
