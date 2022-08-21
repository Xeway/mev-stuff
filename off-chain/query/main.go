package query

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
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

type ResStruct struct {
	Data struct {
		Pairs []Pair `json:"pairs"`
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

	data, err := ioutil.ReadAll(rawRes.Body)
	if err != nil {
		panic(err)
	}

	var res ResStruct
	json.Unmarshal(data, &res)

	return res.Data.Pairs
}
