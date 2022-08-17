package main

import (
	"context"
	"fmt"
	"time"

	"github.com/Xeway/mev-stuff/modules"
	"github.com/ethereum/go-ethereum/ethclient"
)

func EvaluateArb(client *ethclient.Client) {
	amount := int64(1)

	graph := modules.GetAllRates(client, amount)
	bestPath := modules.FindBestPath(graph)

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
