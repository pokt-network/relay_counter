module github.com/pokt-network/relay_counter

go 1.13

require (
	github.com/pokt-network/pocket-core v0.0.0-20220420164902-8ad860a86be1
	github.com/tendermint/tendermint v0.33.7
)

replace github.com/tendermint/tendermint => github.com/pokt-network/tendermint v0.32.11-0.20220420160934-de1729fc7dba // indirect

replace github.com/tendermint/tm-db => github.com/pokt-network/tm-db v0.5.2-0.20220118210553-9b2300f289ba
