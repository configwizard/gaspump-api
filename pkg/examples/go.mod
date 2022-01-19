module github.com/nspcc-dev/neofs-sdk-go/client/examples

go 1.16

require (
	github.com/nspcc-dev/neo-go v0.98.0
	github.com/nspcc-dev/neofs-sdk-go v0.0.0-20220114114829-49a17a715986
)

// copied this approach from github.com/grpc/grpc-go/examples
replace github.com/nspcc-dev/neofs-sdk-go => ../../
