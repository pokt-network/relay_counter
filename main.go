package main

import (
	"flag"
	"log"
	"time"
)

func main() {
	now := time.Now().AddDate(0, 0, -1).Format("01-02-06T15:04:05")
	configFilePath := flag.String("config", "config/config.json", "config file path")
	resultFilePath := flag.String("results", "result/"+now+".json", "results file path")
	timelineStart := flag.Int64("timelineStart", -99999, "override timeline.start.")
	timelineEnd := flag.Int64("timelineEnd", -99999, "override timeline.end.")
	timelineUnit := flag.String("timelineUnit", "", "override timeline.unit.")

	// node
	endpoint := flag.String("endpoint", "", "override endpoint.")
	httpRetry := flag.Int("httpRetry", -1, "override http_retry.")

	// params
	blocksPerSession := flag.Int64("blocksPerSession", -1, "override params.blocks_per_session.")
	blockTimeInMin := flag.Int64("blockTimeInMin", -1, "override params.approx_block_time_in_min.")
	flag.Parse()

	log.Println("Attempting to read Config file:")
	c := getConfig(*configFilePath)

	// parse config file partial overrides
	c = overrideConfig(
		c,
		*timelineStart, *timelineEnd, *timelineUnit,
		*endpoint, *httpRetry,
		*blocksPerSession, *blockTimeInMin,
	)

	log.Println("Config Processed:")
	log.Println(c)

	log.Println("Testing Pocket Endpoint")
	if err := testEndpoint(c.Endpoint); err != nil {
		log.Fatal(err)
	}
	log.Println("Converting Timeline To Block Heights")
	timelineReport, err := ConvertTimelineToHeights(c)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Beginning to retrieve the transactions and claims from the blockchain")
	blockTxsMap, claimsMap, startSupply, endSupply := GetChainData(timelineReport.MinHeight, timelineReport.MaxHeight, c)
	log.Println("Creating a report from the blockchain data")
	result := ProcessChainData(blockTxsMap, claimsMap, startSupply, endSupply, timelineReport)
	log.Println("Writing the result to a report file under /result/<date>.json")
	writeResultFile(result, *resultFilePath)
	log.Println("Done")
}
