package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/antihax/optional"
	"github.com/gateio/gateapi-go/v5"
	"github.com/shopspring/decimal"
	"log"
	"math/rand"
	"time"
)

func SpotDemo(config *RunConfig) {
	client := gateapi.NewAPIClient(gateapi.NewConfiguration())
	// Setting host is optional. It defaults to https://api.gateio.ws/api/v4
	client.ChangeBasePath(config.BaseUrl)
	ctx := context.WithValue(context.Background(), gateapi.ContextGateAPIV4, gateapi.GateAPIV4{
		Key:    config.ApiKey,
		Secret: config.ApiSecret,
	})

	pairs, _, err := client.SpotApi.ListCurrencyPairs(ctx)
	if err != nil {
		log.Fatal(err)
	}
	for _, v := range pairs {
		if v.Base == "GAS" || v.Quote == "GAS" {
			 log.Printf("found %+v\r\n", v)
		}
	}
	//log.Printf("pairs %+v\r\n", pairs)
	currencyPair := "GAS_USDT"
	currency := "USDT"
	cp, _, err := client.SpotApi.GetCurrencyPair(ctx, currencyPair)
	if err != nil {
		panicGateError(err)
	}
	logger.Printf("testing against currency pair: %s\n", cp.Id)
	minAmount := cp.MinQuoteAmount

	tickers, _, err := client.SpotApi.ListTickers(ctx, &gateapi.ListTickersOpts{CurrencyPair: optional.NewString(cp.Id)})
	if err != nil {
		panicGateError(err)
	}
	lastPrice := tickers[0].Last

	log.Println("converting", minAmount)
	// better avoid using float, take the following decimal library for example
	// `go get github.com/shopspring/decimal`
	orderAmount := decimal.RequireFromString(minAmount)
	orderAmount = orderAmount.Mul(decimal.NewFromInt32(2))

	balance, _, err := client.SpotApi.ListSpotAccounts(ctx, &gateapi.ListSpotAccountsOpts{Currency: optional.NewString(currency)})
	if err != nil {
		panicGateError(err)
	}
	if decimal.RequireFromString(balance[0].Available).Cmp(orderAmount) < 0 {
		logger.Fatal("balance not enough")
	}

	newOrder := gateapi.Order{
		Text:         "t-my-custom-id", // optional custom order ID
		CurrencyPair: cp.Id,
		Type:         "limit",
		Account:      "spot", // create spot order. set to "margin" if creating margin orders
		Side:         "buy",
		Amount:       orderAmount.String(),
		Price:        lastPrice, // use last price
		TimeInForce:  "gtc",
		AutoBorrow:   false,
	}
	logger.Printf("place a spot %s order in %s with amount %s and price %s\n", newOrder.Side, newOrder.CurrencyPair, newOrder.Amount, newOrder.Price)
	createdOrder, _, err := client.SpotApi.CreateOrder(ctx, newOrder)
	if err != nil {
		panicGateError(err)
	}
	logger.Printf("order created with ID: %s, status: %s\n", createdOrder.Id, createdOrder.Status)
	if createdOrder.Status == "open" {
		order, _, err := client.SpotApi.GetOrder(ctx, createdOrder.Id, createdOrder.CurrencyPair)
		if err != nil {
			panicGateError(err)
		}
		logger.Printf("order %s filled: %s, left: %s\n", order.Id, order.FilledTotal, order.Left)
		result, _, err := client.SpotApi.CancelOrder(ctx, createdOrder.Id, createdOrder.CurrencyPair)
		if err != nil {
			panicGateError(err)
		}
		if result.Status == "cancelled" {
			logger.Printf("order %s cancelled\n", createdOrder.Id)
		}
	} else {
		// order finished
		trades, _, err := client.SpotApi.ListMyTrades(ctx, createdOrder.CurrencyPair,
			&gateapi.ListMyTradesOpts{OrderId: optional.NewString(createdOrder.Id)})
		if err != nil {
			panicGateError(err)
		}
		for _, t := range trades {
			logger.Printf("order %s filled %s with price: %s\n", t.OrderId, t.Amount, t.Price)
		}
	}
}

var logger = log.New(flag.CommandLine.Output(), "", log.LstdFlags)

func panicGateError(err error) {
	if e, ok := err.(gateapi.GateAPIError); ok {
		log.Fatal(fmt.Sprintf("Gate API error, label: %s, message: %s", e.Label, e.Message))
	}
	log.Fatal(err)
}


/* testnet gate.io
 * api key ab7daf787cc5b79523ed20fe5ca38d43
 * secret f8f966f65ab785fb05d5480a41170cbd9dd4e2538667793f325e63b1adfcffbf
 */

func withdraw(config *RunConfig) {
	client := gateapi.NewAPIClient(gateapi.NewConfiguration())
	// Setting host is optional. It defaults to https://api.gateio.ws/api/v4
	client.ChangeBasePath(config.BaseUrl)
	ctx := context.WithValue(context.Background(), gateapi.ContextGateAPIV4, gateapi.GateAPIV4{
		Key:    config.ApiKey,
		Secret: config.ApiSecret,
	})
	ledgerRecord := gateapi.LedgerRecord{
		Amount:    "1",
		Currency:  "GAS",
		Address:   "recipient address",
		Memo:      "gaspump-credit-account",
	} // LedgerRecord -

	result, _, err := client.WithdrawalApi.Withdraw(ctx, ledgerRecord)
	if err != nil {
		if e, ok := err.(gateapi.GateAPIError); ok {
			fmt.Printf("gate api error: %s\n", e.Error())
		} else {
			fmt.Printf("generic error: %s\n", err.Error())
		}
	} else {
		fmt.Println(result)
	}
}
//  // client.ChangeBasePath("https://fx-api-testnet.gateio.ws/api/v4")
func main() {
	var key, secret, baseUrl string
	//flag.StringVar(&key, "k", "ab7daf787cc5b79523ed20fe5ca38d43", "Gate APIv4 key")
	//flag.StringVar(&secret, "s", "f8f966f65ab785fb05d5480a41170cbd9dd4e2538667793f325e63b1adfcffbf", "Gate APIv4 secret")
	//flag.StringVar(&baseUrl, "u", "", "API based URL used")
	//flag.Parse()

	key = "5ff790cd52e3cbbc4730fd7f9921fc22"
	secret = "93e9b42004a46fdc422f5fec85d05810357765a6010e7efa97b6c2d12c348f2d"
	baseUrl = "https://api.gateio.ws/api/v4"
	//usage := fmt.Sprintf("Usage: %s -k <api-key> -s <api-secret> <spot|margin|futures>", os.Args[0])

	//if key == "" || secret == "" {
	//	logger.Println(usage)
	//	flag.PrintDefaults()
	//	os.Exit(1)
	//}
	//if flag.NArg() < 1 {
	//	logger.Println(usage)
	//	flag.PrintDefaults()
	//	os.Exit(1)
	//}

	runConfig, err := NewRunConfig(key, secret, &baseUrl)
	if err != nil {
		logger.Fatal(err)
	}
	rand.Seed(time.Now().Unix())
	SpotDemo(runConfig)

}
