module github.com/pokt-network/relay_counter

go 1.13

require (
	github.com/dgraph-io/badger/v2 v2.2007.2 // indirect
	github.com/mitchellh/go-ps v1.0.0 // indirect
	github.com/pokt-network/pocket-core v0.0.0-20210429190449-f794bc74b167
	github.com/tendermint/go-amino v0.15.0 // indirect
	github.com/tendermint/tendermint v0.33.7
)

replace github.com/tendermint/tendermint => github.com/pokt-network/tendermint v0.32.11-0.20210427155510-04e1c67f3eed // indirect
