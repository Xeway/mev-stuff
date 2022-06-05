import { ethers } from "ethers";
import dotenv from "dotenv";
dotenv.config();

async function main() {
    const provider = new ethers.providers.WebSocketProvider(`wss://eth-mainnet.alchemyapi.io/v2/${process.env.WEBSOCKET_ALCHEMY_MAINNET}`);
    
    provider.on("pending", async (tx) => {
        const txInfo = await provider.getTransaction(tx);
        console.log(txInfo);
    });
}

main();
