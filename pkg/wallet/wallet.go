package wallet

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"github.com/nspcc-dev/neo-go/cli/flags"
	"github.com/nspcc-dev/neo-go/pkg/core/transaction"
	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	"github.com/nspcc-dev/neo-go/pkg/rpc/client"
	"github.com/nspcc-dev/neo-go/pkg/util"
	"github.com/nspcc-dev/neo-go/pkg/wallet"
	"strconv"
)

type RPC_NETWORK string
const (
	RPC_TESTNET RPC_NETWORK = "http://seed1t4.neo.org:20332"
	RPC_MAINNET RPC_NETWORK = "http://seed1t4.neo.org:20332"
)

func GenerateNewWallet(path string) (*wallet.Wallet, error) {
	acc, err := wallet.NewAccount()
	if err != nil {
		return &wallet.Wallet{}, err
	}
	w, err := wallet.NewWallet(path)
	w.AddAccount(acc)
	return w, err
}

func GenerateNewSecureWallet(path, name, password string) (*wallet.Wallet, error) {
	w, err := wallet.NewWallet(path)
	w.CreateAccount(name, password)
	return w, err
}

func RetrieveWallet(path string) (*wallet.Wallet, error) {
	w, err := wallet.NewWalletFromFile(path)
	if err != nil {
		return nil, fmt.Errorf("can't read the wallet: %walletPath", err)
	}
	return w, nil
}
func GetCredentialsFromWallet(address, password string, w *wallet.Wallet) (*ecdsa.PrivateKey, error) {
	return getKeyFromWallet(w, address, password)
}
func GetCredentialsFromPath(path, address, password string) (*ecdsa.PrivateKey, error) {
	w, err := wallet.NewWalletFromFile(path)
	if err != nil {
		return nil, fmt.Errorf("can't read the wallet: %walletPath", err)
	}

	return getKeyFromWallet(w, address, password)
}
func UnlockWallet(path, address, password string) (*wallet.Account, error) {
	w, err := wallet.NewWalletFromFile(path)
	if err != nil {
		return nil, err
	}
	var addr util.Uint160
	if len(address) == 0 {
		addr = w.GetChangeAddress()
	} else {
		addr, err = flags.ParseAddress(address)
		if err != nil {
			return nil, fmt.Errorf("invalid address")
		}
	}

	acc := w.GetAccount(addr)
	err = acc.Decrypt(password, w.Scrypt)
	if err != nil {
		return nil, err
	}
	return acc, nil
}
type Nep17Tokens struct {
	Asset util.Uint160 `json:"asset"`
	Amount uint64 `json:"amount""`
	Symbol string `json:"symbol"`
	Info wallet.Token `json:"meta"`
	Error error `json:"error"`
}
func GetNep17Balances(walletAddress string, network RPC_NETWORK) (map[string]Nep17Tokens, error) {
	ctx := context.Background()
	// use endpoint addresses of public RPC nodes, e.g. from https://dora.coz.io/monitor
	cli, err := client.New(ctx, string(network), client.Options{})
	if err != nil {
		return map[string]Nep17Tokens{}, err
	}
	err = cli.Init()
	if err != nil {
		return map[string]Nep17Tokens{}, err
	}
	recipient, err := StringToUint160(walletAddress)
	if err != nil {
		return map[string]Nep17Tokens{}, err
	}
	balances, err := cli.GetNEP17Balances(recipient)
	tokens := make(map[string]Nep17Tokens)
	for _, v := range balances.Balances {
		tokInfo := Nep17Tokens{}
		symbol, err := cli.NEP17Symbol(v.Asset)
		if err != nil {
			tokInfo.Error = err
			continue
		}
		tokInfo.Symbol = symbol
		fmt.Println(v.Asset, v.Asset)
		number, err := strconv.ParseUint(v.Amount, 10, 64)
		if err != nil {
			tokInfo.Error = err
			continue
		}
		tokInfo.Amount = number

		info, err := cli.NEP17TokenInfo(v.Asset)
		if err != nil {
			tokInfo.Error = err
			continue
		}
		tokInfo.Info = *info
		tokens[symbol] = tokInfo
	}

	return tokens, nil
}
//TransferToken transfer Nep17 token to another wallet, for instance use address here https://testcdn.fs.neo.org/doc/integrations/endpoints/
//simple example https://gist.github.com/alexvanin/4f22937b99990243a60b7abf68d7458c
func TransferToken(a *wallet.Account, amount int64, walletTo string, token util.Uint160, network RPC_NETWORK) (string, error) {
	ctx := context.Background()
	// use endpoint addresses of public RPC nodes, e.g. from https://dora.coz.io/monitor
	cli, err := client.New(ctx, string(network), client.Options{})
	if err != nil {
		return "", err
	}
	err = cli.Init()
	if err != nil {
		return "", err
	}
	recipient, err := StringToUint160(walletTo)
	if err != nil {
		return "", err
	}
	transaction.NewTransactionFromBytes()
	cli.Tra
	txHash, err := cli.TransferNEP17(a, recipient, token, amount, 0, nil, nil)
	le := txHash.StringLE()
	return le, err
}

// getKeyFromWallet fetches private key from neo-go wallet structure
func getKeyFromWallet(w *wallet.Wallet, addrStr, password string) (*ecdsa.PrivateKey, error) {
	var (
		addr util.Uint160
		err  error
	)

	if addrStr == "" {
		addr = w.GetChangeAddress()
	} else {
		addr, err = flags.ParseAddress(addrStr)
		if err != nil {
			return nil, fmt.Errorf("invalid wallet address %s: %w", addrStr, err)
		}
	}

	acc := w.GetAccount(addr)
	if acc == nil {
		return nil, fmt.Errorf("invalid wallet address %s: %w", addrStr, err)
	}

	if err := acc.Decrypt(password, keys.NEP2ScryptParams()); err != nil {
		return nil, errors.New("[decrypt] invalid password - " + err.Error())

	}

	return &acc.PrivateKey().PrivateKey, nil
}


