package main
import "C"
import (
	"context"
	"github.com/configwizard/gaspump-api/pkg/wallet"
	"github.com/nspcc-dev/neofs-sdk-go/client"
	"sync"
)

var mtx sync.Mutex

//export Test
func Test() string {
	return "alex walker"
}
//export RetrieveNeoFSBalance
func RetrieveNeoFSBalance(walletPath, password string) (int64, error) {
	ctx := context.Background()
	// First obtain client credentials: private key of request owner
	key, err := wallet.GetCredentialsFromPath(walletPath, "", password)
	if err != nil {
		return 0, err
	}
	defer mtx.Unlock()
	cli, err := client.New(
		// provide private key associated with request owner
		client.WithDefaultPrivateKey(key),
		// find endpoints in https://testcdn.fs.neo.org/doc/integrations/endpoints/
		client.WithURIAddress("grpcs://st01.testnet.fs.neo.org:8082", nil),
		// check client errors in go compatible way
		client.WithNeoFSErrorParsing(),
	)
	if err != nil {
		return 0, err
	}

	owner, err := wallet.OwnerIDFromPrivateKey(key)
	if err != nil {
		return 0, err
	}
	get := client.PrmBalanceGet{}
	get.SetAccount(*owner)
	result, err := cli.BalanceGet(ctx, get)
	return result.Amount().Value(), err
}
func main() {}
