package main

import (
	"log"
)

func main() {
	log.Println("Attempting to read Config file:")
	c := getConfig()
	log.Println("Config File Processed:")
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
	writeResultFile(result)
	log.Println("Done")
}
