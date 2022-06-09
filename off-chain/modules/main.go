package modules

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"

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

func GetAllUniswapPairs(client *ethclient.Client, routerAddresses []string) map[common.Address][]*big.Int {
	// this map is structured like that : pairs[address of stablecoin like USDC] = [return amount of stablecoin of exchange A, return amount of stablecoin of exchange B...]
	pairs := make(map[common.Address][]*big.Int)

	executorWallet := common.HexToAddress(addresses.EXECUTOR_WALLET)
	options := &bind.CallOpts{true, executorWallet, nil, context.Background()}
	wethAddress := common.HexToAddress(addresses.WETH_ADDRESS)

	for _, routerAddress := range routerAddresses {
		routerAddress := common.HexToAddress(routerAddress)

		instanceRouter, err := uniswap_router.NewUniswapRouterCaller(routerAddress, client)

		if err != nil {
			log.Fatal(err)
		}

		for _, stableAddress := range addresses.STABLE_ADDRESSES {
			stableAddress := common.HexToAddress(stableAddress)

			stableInstance, err := erc_20.NewErc20Caller(stableAddress, client)
			if err != nil {
				log.Fatal(err)
			}

			stableDecimals, err := stableInstance.Decimals(options)
			if err != nil {
				log.Fatal(err)
			}

			stableAmount := 1000

			amountIn := big.NewInt(0).Mul(big.NewInt(int64(stableAmount)), big.NewInt(0).Exp(big.NewInt(10), big.NewInt(int64(stableDecimals)), nil))

			amountOut, err := instanceRouter.GetAmountsOut(options, amountIn, []common.Address{stableAddress, wethAddress})
			if err != nil {
				log.Fatal(err)
			}

			pairs[stableAddress] = append(pairs[stableAddress], amountOut[1])
		}
	}

	return pairs
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
