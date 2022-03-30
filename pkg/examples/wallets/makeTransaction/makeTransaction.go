package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/nspcc-dev/neo-go/pkg/rpc/client"
	"github.com/nspcc-dev/neo-go/pkg/smartcontract"
	"math/big"
	"time"

	//"github.com/configwizard/gaspump-api/pkg/examples/utils"
	"github.com/configwizard/gaspump-api/pkg/wallet"
	"io/ioutil"
	"log"
	"os"
)

const usage = `Example

$ ./balance -wallets ./sample_wallets/wallet.json
password is password
`

var (
	walletPath = flag.String("wallets", "./pkg/examples/sample_wallets/wallet.json", "path to JSON wallets file")
	walletAddr = flag.String("address", "", "wallets address [optional]")
	createWallet = flag.Bool("create", false, "create a wallets")
	password = flag.String("password", "", "wallet password")
)

func main() {
	path, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}
	fmt.Println(path)  // for example /home/user
	flag.Usage = func() {
		_, _ = fmt.Fprintf(os.Stderr, usage)
		flag.PrintDefaults()
	}
	flag.Parse()

	ctx := context.Background()

	if *createWallet {
		secureWallet, err := wallet.GenerateNewSecureWallet(*walletPath, "some account label", *password)
		if err != nil {
			log.Fatal("error generating wallets", err)
		}
		file, _ := json.MarshalIndent(secureWallet, "", " ")
		_ = ioutil.WriteFile(*walletPath, file, 0644)
		log.Printf("created new wallets\r\n%+v\r\n", file)
		os.Exit(0)
	}

	cli, err := client.New(ctx, string(wallet.RPC_TESTNET), client.Options{})
	if err != nil {
		log.Fatal(err)
	}
	err = cli.Init()
	if err != nil {
		log.Fatal(err)
	}

	acc, err := wallet.UnlockWallet(*walletPath, "", *password)
	if err != nil {
		log.Fatal("can't unlock wallet:", err)
	}

	account, err := wallet.StringToUint160(acc.Address)
	param := smartcontract.Parameter{
		Type:  smartcontract.Hash160Type,
		Value: account,
	}
	params := []smartcontract.Parameter{param}
	operation := "balanceOf" //symbol
	network := wallet.RPC_TESTNET
	transactionID, transaction, err := wallet.CreateTransactionFromFunctionCall("0x0a81b80376a65003781f140d1b87b6531f706215", operation, network, acc, params)
	if err != nil {
		log.Fatal("transaction failed ", err)
	}
	log.Printf("success: transaction %+v\r\nSleeping...", transaction)
	time.Sleep(20 * time.Second) //wait for the network to process transaction
	applicationLog, err := wallet.GetLogForTransaction(wallet.RPC_TESTNET, transactionID)
	for _, v := range applicationLog.Executions {
		log.Printf("v %+v\r\n", v)
		for i, k := range v.Stack {
			switch v := k.Value().(type) {
			case *big.Int:
				log.Printf("[int] stack item %d %d\r\n", i, k.Value().(*big.Int))
			case string:
				log.Printf("[string] stack item %d %s\r\n", i, k.Value().(string))
			case []byte:
				log.Printf("[[]byte] stack item %d %s\r\n", i, string(k.Value().([]byte)))
			default:
				fmt.Printf("I don't know about type %T!\n", v)
			}

		}
	}
	fmt.Printf("err %s\r\nappl %+v\r\n, tID %s\r\n", err, applicationLog, transactionID.StringLE())
}


