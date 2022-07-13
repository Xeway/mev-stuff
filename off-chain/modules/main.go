package modules

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math"
	"math/big"
	"strconv"
	"sync"

	erc_20 "github.com/Xeway/mev-stuff/abi/go/erc_20"
	uniswap_router "github.com/Xeway/mev-stuff/abi/go/uniswap_router"
	"github.com/Xeway/mev-stuff/addresses"
	Models "github.com/Xeway/mev-stuff/models"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

type ExchangeAndAmount struct {
	ExchangeAddr common.Address
	AmountOut    *big.Int
}

func GetAllRates(client *ethclient.Client, amount int64) [][]*big.Int {
	executorWallet := common.HexToAddress(addresses.EXECUTOR_WALLET)
	options := &bind.CallOpts{true, executorWallet, nil, context.Background()}

	routerAddress := common.HexToAddress(addresses.UNISWAP_ROUTER_ADDRESS)
	instanceRouter, err := uniswap_router.NewUniswapRouterCaller(routerAddress, client)

	if err != nil {
		log.Fatal(err)
	}

	graph := make([][]*big.Int, len(addresses.TOKEN_ADDRESSES))

	var wg1 sync.WaitGroup
	wg1.Add(len(addresses.TOKEN_ADDRESSES))

	for i, src := range addresses.TOKEN_ADDRESSES {
		go func(i int, src string) {
			tokenAddrSrc := common.HexToAddress(src)

			graph[i] = make([]*big.Int, len(addresses.TOKEN_ADDRESSES))

			tokenInstance, err := erc_20.NewErc20Caller(tokenAddrSrc, client)
			if err != nil {
				log.Fatal(err)
			}

			tokenDecimals, err := tokenInstance.Decimals(options)
			if err != nil {
				log.Fatal(err)
			}

			amount := big.NewInt(0).Mul(big.NewInt(int64(amount)), big.NewInt(0).Exp(big.NewInt(10), big.NewInt(int64(tokenDecimals)), nil))

			var wg2 sync.WaitGroup
			wg2.Add(len(addresses.TOKEN_ADDRESSES))

			for j, dest := range addresses.TOKEN_ADDRESSES {
				go func(j int, dest string) {
					tokenAddrDest := common.HexToAddress(dest)

					var rate *big.Int

					if tokenAddrSrc == tokenAddrDest {
						rate = big.NewInt(1)
					} else {
						res, err := instanceRouter.GetAmountsOut(options, amount, []common.Address{tokenAddrSrc, tokenAddrDest})
						if err != nil {
							log.Fatal(err)
						}

						rate = res[1]
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

func isOverflow(bigNum *big.Int) bool {
	return bigNum.String() != strconv.Itoa(int(bigNum.Int64()))
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
					isOf := isOverflow(rate)

					if isOf {
						num := make([]*big.Int, 0)

						firstIteration := true
						for isOf {
							if firstIteration {
								num = append(num, big.NewInt(0).Sqrt(rate))
								firstIteration = false
							} else {
								num = append(num, big.NewInt(0).Sqrt(num[len(num)-1]))
							}
							isOf = isOverflow(num[len(num)-1])
						}

						numMul := math.Pow(2, float64(len(num)))

						var bigNumLog float64

						var wg3 sync.WaitGroup
						wg3.Add(int(numMul))

						for k := 0; k < int(numMul); k++ {
							go func() {
								bigNumLog += math.Log10(float64(num[len(num)-1].Int64()))

								wg3.Done()
							}()
						}
						wg3.Wait()

						newGraph[i][j] = -bigNumLog
					} else {
						newGraph[i][j] = -math.Log10(float64(rate.Int64()))
					}

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

func checkIfPresentInArray(pre int, printCycle []int) bool {
	c := make(chan bool)

	counter := len(printCycle)

	for _, val := range printCycle {
		go func(val int, counter *int) {
			if val == pre {
				c <- true
			}

			*counter--

			if *counter == 0 {
				close(c)
			}
		}(val, &counter)
	}

	for range c {
		return false
	}

	return true
}

func ReverseArray(arr []int) []int {
	for i, j := 0, len(arr)-1; i < j; i, j = i+1, j-1 {
		arr[i], arr[j] = arr[j], arr[i]
	}

	return arr
}

func FindBestPath(graph [][]*big.Int) []string {
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

				for checkIfPresentInArray(pre[i], printCycle) {
					printCycle = append(printCycle, pre[i])
					i = pre[i]
				}
				printCycle = append(printCycle, pre[i])

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

	bestPathAddr := make([]string, len(bestPath))
	for _, v := range bestPath {
		bestPathAddr = append(bestPathAddr, addresses.TOKEN_ADDRESSES[v])
	}

	return bestPathAddr
}

func GetAllUniswapAmountOut(client *ethclient.Client, stableAmount int) map[common.Address][]ExchangeAndAmount {
	// this map is structured like that : amountsOut[address of stablecoin like USDC] = an struct that contain the exchange + the amount out
	amountsOut := make(map[common.Address][]ExchangeAndAmount)

	executorWallet := common.HexToAddress(addresses.EXECUTOR_WALLET)
	options := &bind.CallOpts{true, executorWallet, nil, context.Background()}
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

func GetLatestBlock(client ethclient.Client) *Models.Block {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()

	// query the latest block
	header, _ := client.HeaderByNumber(context.Background(), nil)
	blockNumber := big.NewInt(header.Number.Int64())
	block, err := client.BlockByNumber(context.Background(), blockNumber)

	if err != nil {
		log.Fatal(err)
	}

	// Build the response to our model
	_block := &Models.Block{
		BlockNumber:       block.Number().Int64(),
		Timestamp:         block.Time(),
		Difficulty:        block.Difficulty().Uint64(),
		Hash:              block.Hash().String(),
		TransactionsCount: len(block.Transactions()),
		Transactions:      []Models.Transaction{},
	}

	for _, tx := range block.Transactions() {
		_block.Transactions = append(_block.Transactions, Models.Transaction{
			Hash:     tx.Hash().String(),
			Value:    tx.Value().String(),
			Gas:      tx.Gas(),
			GasPrice: tx.GasPrice().Uint64(),
			Nonce:    tx.Nonce(),
			To:       tx.To().String(),
		})
	}

	return _block
}

func GetTxByHash(client ethclient.Client, hash common.Hash) *Models.Transaction {

	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()

	tx, pending, err := client.TransactionByHash(context.Background(), hash)
	if err != nil {
		fmt.Println(err)
	}

	return &Models.Transaction{
		Hash:     tx.Hash().String(),
		Value:    tx.Value().String(),
		Gas:      tx.Gas(),
		GasPrice: tx.GasPrice().Uint64(),
		To:       tx.To().String(),
		Pending:  pending,
		Nonce:    tx.Nonce(),
	}
}

func TransferEth(client ethclient.Client, privKey string, to string, amount int64) (string, error) {

	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()

	// Assuming you've already connected a client, the next step is to load your private key.
	privateKey, err := crypto.HexToECDSA(privKey)
	if err != nil {
		return "", err
	}

	// Function requires the public address of the account we're sending from -- which we can derive from the private key.
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return "", err
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	// Now we can read the nonce that we should use for the account's transaction.
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return "", err
	}

	value := big.NewInt(amount) // in wei (1 eth)
	gasLimit := uint64(21000)   // in units
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return "", err
	}

	// We figure out who we're sending the ETH to.
	toAddress := common.HexToAddress(to)
	var data []byte

	// We create the transaction payload
	tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, data)

	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		return "", err
	}

	// We sign the transaction using the sender's private key
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		return "", err
	}

	// Now we are finally ready to broadcast the transaction to the entire network
	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return "", err
	}

	// We return the transaction hash
	return signedTx.Hash().String(), nil
}

// GetAddressBalance returns the given address balance =P
func GetAddressBalance(client ethclient.Client, address string) (string, error) {
	account := common.HexToAddress(address)
	balance, err := client.BalanceAt(context.Background(), account, nil)
	if err != nil {
		return "0", err
	}

	return balance.String(), nil
}
