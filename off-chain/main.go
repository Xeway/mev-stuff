package main

import (
	"fmt"
	"log"
	"net/http"

	Handlers "github.com/Xeway/mev-stuff/handler"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gorilla/mux"
)

func main() {
	// Create a client instance to connect to our provider
	client, err := ethclient.Dial("https://eth-mainnet.alchemyapi.io/v2/YgladcqcW2iHbKZVTYd4S0VFcESU4UKw")

	if err != nil {
		fmt.Println(err)
	}

	// Create a mux router
	r := mux.NewRouter()

	// We will define a single endpoint
	r.Handle("/api/v1/eth/{module}", Handlers.ClientHandler{client})
	log.Fatal(http.ListenAndServe(":8080", r))
}
