package main

import (
	"fmt"
	"github.com/amlwwalker/gaspump-api/pkg/wallet"
	"os"
)

func main() {
	peers, err := wallet.GetPeers(wallet.RPC_TESTNET)
	if err != nil {
		fmt.Println("error retreiving peers", err)
		os.Exit(2)
	}
	fmt.Printf("peers %+v\r\n", peers)
}
