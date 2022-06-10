package modules

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"
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

func GetAllUniswapAmountOut(client *ethclient.Client, stableAmount int) map[common.Address][](map[common.Address]*big.Int) {
	// this map is structured like that : amountsOut[address of stablecoin like USDC] = an array of the amount out according to the exchange
	amountsOut := make(map[common.Address][](map[common.Address]*big.Int))

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

					exchangeToAmount := make(map[common.Address]*big.Int)
					exchangeToAmount[routerAddress] = amountOut[1]

					amountsOut[stableAddress] = append(amountsOut[stableAddress], exchangeToAmount)

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
	stable        common.Address
	bestAmount    *big.Int
	bestExchange  common.Address
	worstAmount   *big.Int
	worstExchange common.Address
}

func GetBestArbitrageOpportunity(client *ethclient.Client, amountsOut map[common.Address][](map[common.Address]*big.Int)) Opportunity {
	var bestOpportunity Opportunity

	bestDelta := big.NewInt(0)

	for stableAddr, exchangeToAmount := range amountsOut { // we get every array ordered by exchange according to each stable coin
		bestAmount := big.NewInt(0)
		var bestExchange common.Address

		worstAmount := big.NewInt(0)
		var worstExchange common.Address

		for i, mapE2A := range exchangeToAmount { // we get every map exchange => amount
			for exchangeAddr, amount := range mapE2A { // we get the exchange address and the amount
				if amount.Cmp(bestAmount) == 1 {
					bestAmount = amount
					bestExchange = exchangeAddr
				}

				if amount.Cmp(worstAmount) == -1 || i == 0 {
					worstAmount = amount
					worstExchange = exchangeAddr
				}
			}
		}

		amountDelta := big.NewInt(0).Sub(bestAmount, worstAmount)

		if amountDelta.Cmp(bestDelta) == 1 {
			bestDelta = amountDelta

			bestOpportunity = Opportunity{
				stable:        stableAddr,
				bestAmount:    bestAmount,
				bestExchange:  bestExchange,
				worstAmount:   worstAmount,
				worstExchange: worstExchange,
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
