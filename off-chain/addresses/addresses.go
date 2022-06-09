package addresses

var (
	UNISWAP_ROUTER_ADDRESS   string = "0x7a250d5630B4cF539739dF2C5dAcb4c659F2488D"
	SUSHISWAP_ROUTER_ADDRESS string = "0xd9e1cE17f2641f24aE83637ab66a2cca9C378B9F"
	CRO_ROUTER_ADDRESS       string = "0xCeB90E4C17d626BE0fACd78b79c9c87d7ca181b3"
	// ZEUS_ROUTER_ADDRESS      string = "0x5389c0467010E318B581990d4a8222979fC001d2"
	LUA_ROUTER_ADDRESS string = "0x1d5C6F1607A171Ad52EFB270121331b3039dD83e"

	ROUTER_ADDRESSES = []string{
		UNISWAP_ROUTER_ADDRESS,
		SUSHISWAP_ROUTER_ADDRESS,
		CRO_ROUTER_ADDRESS,
		// ZEUS_ROUTER_ADDRESS,
		LUA_ROUTER_ADDRESS,
	}

	WETH_ADDRESS string = "0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2"
	USDT_ADDRESS string = "0xdAC17F958D2ee523a2206206994597C13D831ec7"
	USDC_ADDRESS string = "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48"
	DAI_ADDRESS  string = "0x6B175474E89094C44Da98b954EedeAC495271d0F"

	STABLE_ADDRESSES = []string{
		USDT_ADDRESS,
		USDC_ADDRESS,
		DAI_ADDRESS,
	}

	EXECUTOR_WALLET string = "0xE4E6dC19efd564587C46dCa2ED787e45De17E7E1"
)
