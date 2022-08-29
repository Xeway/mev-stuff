// Arbitrage for the same pair (STABLE-WETH) on different exchanges

package main

import (
	"context"
	"fmt"
	"time"

	"github.com/Xeway/mev-stuff/addresses"
	arb "github.com/Xeway/mev-stuff/arb_utils"
	"github.com/Xeway/mev-stuff/query"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

func QueryBiggestPairs() {
	pairs := query.GetPairsWithMostReserves()

	for i := 0; i < len(pairs); i++ {
		addresses.PAIR_ADDRESSES = append(addresses.PAIR_ADDRESSES, common.HexToAddress(pairs[i].Id))
	}
}

func EvaluateArb(client *ethclient.Client) {
	amount := int64(1)

	amountsOut := arb.GetAllUniswapAmountOut(client, amount)
	bestOpportunity, err := arb.GetBestArbitrageOpportunity(client, amountsOut)
	if err != nil {
		fmt.Println(err)
		EvaluateArb(client)
	}
	fmt.Println(bestOpportunity)
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
