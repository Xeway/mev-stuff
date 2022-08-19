package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type ResStruct struct {
	Data struct {
		Pairs []struct {
			Id     string `json:"id"`
			Token0 struct {
				Id     string `json:"id"`
				Symbol string `json:"symbol"`
			} `json:"token0"`
			Token1 struct {
				Id     string `json:"id"`
				Symbol string `json:"symbol"`
			} `json:"token1"`
			ReserveETH string `json:"reserveETH"`
			ReserveUSD string `json:"reserveUSD"`
			VolumeUSD  string `json:"volumeUSD"`
		} `json:"pairs"`
	} `json:"data"`
}

func main() {
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
						symbol
					}
					token1 {
						id
						symbol
					}
					reserveETH
					reserveUSD
					volumeUSD
				}
			}
		`,
	}

	jsonReq, err := json.Marshal(jsonData)
	if err != nil {
		fmt.Println(err)
	}

	req, err := http.NewRequest(http.MethodPost, "https://api.thegraph.com/subgraphs/name/uniswap/uniswap-v2", bytes.NewBuffer(jsonReq))
	if err != nil {
		fmt.Println(err)
	}

	client := &http.Client{}

	rawRes, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	defer rawRes.Body.Close()

	data, err := ioutil.ReadAll(rawRes.Body)
	if err != nil {
		fmt.Println(err)
	}

	var res ResStruct
	json.Unmarshal(data, &res)

	fmt.Println(res.Data.Pairs)
}
