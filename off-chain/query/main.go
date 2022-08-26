package query

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

type Pair struct {
	Id     string `json:"id"`
	Token0 struct {
		Id string `json:"id"`
	} `json:"token0"`
	Token1 struct {
		Id string `json:"id"`
	} `json:"token1"`
	ReserveETH string `json:"reserveETH"`
}

type Token struct {
	Id       string `json:"id"`
	Decimals string `json:"decimals"`
}

type ResStruct struct {
	Data struct {
		Pairs  []Pair  `json:"pairs"`
		Tokens []Token `json:"tokens"`
	} `json:"data"`
}

func GetPairsWithMostReserves() []Pair {
	jsonData := map[string]string{
		"query": `
			{
				pairs(
					first: 100
					orderBy: reserveETH
					orderDirection: desc
				) {
					id
					token0 {
						id
					}
					token1 {
						id
					}
					reserveETH
				}
			}
		`,
	}

	res := requestTheGraph(jsonData)

	return res.Data.Pairs
}

func GetTokensWithMostVolume() []Token {
	jsonData := map[string]string{
		"query": `
			{
				tokens(
					first: 100
					orderBy: tradeVolumeUSD
					orderDirection: desc
				) {
					id
					decimals
				}
			}
		`,
	}

	res := requestTheGraph(jsonData)

	return res.Data.Tokens
}

func requestTheGraph(jsonData map[string]string) ResStruct {
	jsonReq, err := json.Marshal(jsonData)
	if err != nil {
		panic(err)
	}

	req, err := http.NewRequest(http.MethodPost, "https://api.thegraph.com/subgraphs/name/uniswap/uniswap-v2", bytes.NewBuffer(jsonReq))
	if err != nil {
		panic(err)
	}

	client := &http.Client{}

	rawRes, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer rawRes.Body.Close()

	data, err := io.ReadAll(rawRes.Body)
	if err != nil {
		panic(err)
	}

	var res ResStruct
	json.Unmarshal(data, &res)

	return res
}
