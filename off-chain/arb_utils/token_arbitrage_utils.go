// Functions to perform arbitrages on multiple tokens of the same exchange

package arb

import (
	"context"
	"errors"
	"log"
	"math"
	"math/big"
	"sync"

	bigmath "github.com/Xeway/big-math"
	uniswap_router "github.com/Xeway/mev-stuff/abi/go/uniswap_router"
	"github.com/Xeway/mev-stuff/addresses"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

func GetAllRates(client *ethclient.Client, amount int64) [][]*big.Int {
	executorWallet := common.HexToAddress(addresses.EXECUTOR_WALLET)
	options := &bind.CallOpts{Pending: true, From: executorWallet, BlockNumber: nil, Context: context.Background()}

	routerAddress := common.HexToAddress(addresses.UNISWAP_ROUTER_ADDRESS)
	instanceRouter, err := uniswap_router.NewUniswapRouterCaller(routerAddress, client)

	if err != nil {
		log.Fatal(err)
	}

	graph := make([][]*big.Int, len(addresses.TOKEN_ADDRESSES))

	var wg1 sync.WaitGroup
	wg1.Add(len(addresses.TOKEN_ADDRESSES))

	for i, src := range addresses.TOKEN_ADDRESSES {
		go func(i int, src addresses.Token) {
			graph[i] = make([]*big.Int, len(addresses.TOKEN_ADDRESSES))

			amount := big.NewInt(0).Mul(big.NewInt(int64(amount)), big.NewInt(0).Exp(big.NewInt(10), big.NewInt(src.Decimals), nil))

			var wg2 sync.WaitGroup
			wg2.Add(len(addresses.TOKEN_ADDRESSES))

			for j, dest := range addresses.TOKEN_ADDRESSES {
				go func(j int, dest addresses.Token) {
					var rate *big.Int

					if src.Address == dest.Address {
						rate = big.NewInt(1)
					} else {
						res, err := instanceRouter.GetAmountsOut(options, amount, []common.Address{src.Address, dest.Address})
						if err != nil {
							rate = big.NewInt(1)
						} else {
							rate = res[1]
						}
					}

					graph[i][j] = rate

					wg2.Done()
				}(j, dest)
			}
			wg2.Wait()

			wg1.Done()
		}(i, src)
	}
	wg1.Wait()

	return graph
}

func parseGraph(graph [][]*big.Int) [][]float64 {
	newGraph := make([][]float64, len(graph))

	var wg1 sync.WaitGroup
	wg1.Add(len(graph))

	for i, tokens := range graph {
		go func(i int, tokens []*big.Int) {
			newGraph[i] = make([]float64, len(tokens))

			var wg2 sync.WaitGroup
			wg2.Add(len(tokens))

			for j, rate := range tokens {
				go func(j int, rate *big.Int) {
					newGraph[i][j] = -bigmath.Log10(rate)

					wg2.Done()
				}(j, rate)
			}
			wg2.Wait()

			wg1.Done()
		}(i, tokens)
	}
	wg1.Wait()

	return newGraph
}

func CheckIfNotPresentInArray[G comparable](wanted G, arr []G) bool {
	for _, val := range arr {
		if val == wanted {
			return false
		}
	}

	return true
}

func ReverseArray(arr []int) []int {
	for i, j := 0, len(arr)-1; i < j; i, j = i+1, j-1 {
		arr[i], arr[j] = arr[j], arr[i]
	}

	return arr
}

func FindBestPath(graph [][]*big.Int) ([]common.Address, error) {
	newGraph := parseGraph(graph)

	dist := make([]float64, len(newGraph))
	dist[0] = float64(0)
	for y := 1; y < len(dist); y++ {
		dist[y] = math.Inf(1)
	}

	pre := make([]int, len(newGraph))
	for w := range pre {
		pre[w] = -1
	}

	for z := 0; z < len(newGraph)-1; z++ {
		for i, src := range newGraph {
			for j, dest := range src {
				if dist[i]+dest < dist[j] {
					dist[j] = dist[i] + dest

					pre[j] = i
				}
			}
		}
	}

	var bestPath []int
	bestSum := float64(0)

	for i, src := range newGraph {
		for j, dest := range src {
			if dist[i]+dest < dist[j] {
				printCycle := make([]int, 2)
				printCycle[0] = j
				printCycle[1] = i

				for CheckIfNotPresentInArray(pre[i], printCycle) {
					printCycle = append(printCycle, pre[i])
					i = pre[i]
				}
				printCycle = append(printCycle, pre[i])

				for l := 0; l < len(printCycle); l++ {
					if printCycle[l] == printCycle[len(printCycle)-1] {
						printCycle = printCycle[l:]
						break
					}
				}

				if len(printCycle) <= 3 {
					return []common.Address{}, errors.New("no arbitrage opportunity")
				}

				printCycle = ReverseArray(printCycle)

				var sum float64
				for k := 0; k < len(printCycle)-1; k++ {
					sum += newGraph[printCycle[k]][printCycle[k+1]]
				}
				if sum < bestSum {
					bestSum = sum
					bestPath = printCycle
				}
			}
		}
	}

	bestPathAddr := make([]common.Address, len(bestPath))
	for _, v := range bestPath {
		bestPathAddr = append(bestPathAddr, addresses.TOKEN_ADDRESSES[v].Address)
	}

	return bestPathAddr, nil
}
