package wallet

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"github.com/nspcc-dev/neo-go/cli/flags"
	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	"github.com/nspcc-dev/neo-go/pkg/rpc/client"
	"github.com/nspcc-dev/neo-go/pkg/util"
	"github.com/nspcc-dev/neo-go/pkg/wallet"
)

type RPC_NETWORK string
const (
	RPC_TESTNET RPC_NETWORK = "https://rpc01.testnet.n3.nspcc.ru:21331"
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

//TransferToken transfer to neo fs, for instance use address here https://testcdn.fs.neo.org/doc/integrations/endpoints/
//simple example https://gist.github.com/alexvanin/4f22937b99990243a60b7abf68d7458c
func TransferToken(a *wallet.Account, amount int64, walletTo string, token util.Uint160, network RPC_NETWORK) (util.Uint256, error) {
	ctx := context.Background()
	// use endpoint addresses of public RPC nodes, e.g. from https://dora.coz.io/monitor
	cli, err := client.New(ctx, string(network), client.Options{})
	if err != nil {
		return util.Uint256{}, err
	}
	err = cli.Init()
	if err != nil {
		return util.Uint256{}, err
	}
	recipient, err := stringToUint160(walletTo)
	if err != nil {
		return util.Uint256{}, err
	}
	txHash, err := cli.TransferNEP17(a, recipient, token, amount, 0, nil, nil)
	return txHash, err
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


