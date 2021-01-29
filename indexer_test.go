package main

import (
	"encoding/json"
	"fmt"
	"github.com/pokt-network/pocket-core/crypto"
	"io/ioutil"
	"os"
	"testing"
)

type GatewayApp struct {
	GatewayAAT GatewayAAT`json:"gatewayAAT"`
}

type GatewayAAT struct {
	ApplicationPublicKey string `json:"applicationPublicKey"`
}

func TestGatewayAppsCheck(t *testing.T) {
	f, err := os.OpenFile("result/gateway.json", 0, 0777)
	if err != nil {
		panic(err)
	}
	bz, err := ioutil.ReadAll(f)
	if err != nil {
		panic(err)
	}
	ga := []GatewayApp{}
	err = json.Unmarshal(bz, &ga)
	if err != nil {
		panic(err)
	}
	var total int64
	f2, err := os.OpenFile("result/01-27-21.json", 0, 0777)
	if err != nil {
		panic(err)
	}
	report := Report{}
	bz, err = ioutil.ReadAll(f2)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(bz, &report)
	if err != nil {
		panic(err)
	}
	i:=0
	for _, g := range ga{
		i++
		pk, err := crypto.NewPublicKey(g.GatewayAAT.ApplicationPublicKey)
		if err != nil {
			fmt.Println("Skipping iteration: ", i)
			continue
		}
		for app, rep := range report.AppReports {
			if app == pk.Address().String() {
				total +=rep.TotalRelays
			}
		}
	}
	fmt.Println(total)
}

