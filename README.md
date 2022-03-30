# GasPump Api

The api holds all the backend functionality to interface with NeoFS.

**You can `go get` this library with** `go get github.com/configwizard/gaspump-api`

In effect this is just a wrapper library around the NeoFS SDK. This is not a replacement for it. 

## WARNING

This work is in early days. Don't completely trust it. However please open PRs to help move this along.

This library should probably get more attention if it's going to be relied on in any serious production environment.

## SECOND WARNING

There are parts of this code that are WIP. Mainly around EACL and Bearer tokens. I will improve this area.

## Examples

Run any of the examples, they don't rely on anything except the api and what is contained in each file.

Using the documentation at [neo-doc](https://developers.neo.org/docs/n3/neofs/introduction/concepts/) you should get on just fine.

Infact that documentation was based off of the work here.

0. There is an example to create a new wallet. Once you have a wallet go to the [Neo TestNet Faucet](https://neowish.ngd.network/#/) for Neo and GAS

A couple of examples, then the rest should make sense :)

1. [Retrieve Your NeoFS Balance](/pkg/examples/wallets/retrieveNeoFSBalance/retrieveNeoFSBalance.go)

This example shows you how to use a wallet to get your NeoFS balance.

2. [Transfer A Token](/pkg/examples/wallets/transferToken/transferToken.go)

This example show shows you how to transfer a token from a wallet to another wallet or contract. Specifically the demo defaults to transfering 1 GAS from your wallet to the NeoFS smart contract, giving you credit in which to pay for containers and objects.
After transfering, you could run the previous example again to see your new balance.

3. [Create a container](/pkg/examples/containers/create/createContainer.go)

Once you have transferred GAS to NeoFS contract, you can create a container according to a policy and ACL scheme with this example

4. [Upload an object](/pkg/examples/objects/upload/uploadObject.go)

From the previous example you will have received a container ID. Use that in this example to upload an object to your new container.

## Hosting Privately

### If the repository is private

you will need to be a collaborator on this repository. After that you will need to generate a github access token in your personal developer settings.

Then run

```shell
go env -w GOPRIVATE=github.com/configwizard/gaspump-api
```

then you will need to run

```shell
git config --global url."git@github.com/<username>:<YOUR ACCESS TOKEN>".insteadOf "https://github.com/<username>"
```

replacing `<YOUR ACCESS TOKEN>` with your access token you retrieved earlier.

Now when building a project that uses this api, like gaspump.react you should have no issues building 
