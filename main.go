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

	selector := flag.String("selector", "", "use this to point which method will you use to select block. It can be: timeline (default) or byBlock")

	timelineStart := flag.Int64("timelineStart", -99999, "override timeline.start.")
	timelineEnd := flag.Int64("timelineEnd", -99999, "override timeline.end.")
	timelineUnit := flag.String("timelineUnit", "", "override timeline.unit.")

	startBlock := flag.Int64("startBlock", -99999, "override byBlock.start.")
	endBlock := flag.Int64("endBlock", -99999, "override byBlock.end.")

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
		*selector,
		*timelineStart, *timelineEnd, *timelineUnit,
		*startBlock, *endBlock,
		*endpoint, *httpRetry,
		*blocksPerSession, *blockTimeInMin,
	)

	log.Println("Config Processed:")
	log.Println(c)

	log.Println("Testing Pocket Endpoint")
	if err := testEndpoint(c.Endpoint); err != nil {
		log.Fatal(err)
	}

	blockReport := BlockReport{}

	if c.Selector == "timeline" {
		log.Println("Converting Timeline To Block Heights")
		report, err := ConvertTimelineToHeights(c)
		if err != nil {
			log.Fatal(err)
		}
		blockReport.MinHeight = report.MinHeight
		blockReport.MaxHeight = report.MaxHeight
	} else if c.Selector == "byBlock" {
		log.Println("Using byBlock as block selector.")
		blockReport.MinHeight = c.ByBlock.Start
		blockReport.MaxHeight = c.ByBlock.End
	} else {
		log.Fatal("selector must be one of following: timeline | byBlock")
	}

	log.Println("Beginning to retrieve the transactions and claims from the blockchain")
	blockTxsMap, claimsMap, startSupply, endSupply := GetChainData(blockReport.MinHeight, blockReport.MaxHeight, c)
	log.Println("Creating a report from the blockchain data")
	result := ProcessChainData(blockTxsMap, claimsMap, startSupply, endSupply, c.Selector, blockReport)
	log.Println("Writing the result to a report file under " + *resultFilePath)
	writeResultFile(result, *resultFilePath)
	log.Println("Done")
}
