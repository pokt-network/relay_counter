package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"strings"
)

type Config struct {
	Selector  string   `json:"selector"`
	Timeline  Timeline `json:"timeline"`
	ByBlock   ByBlock  `json:"byBlock"`
	Endpoint  string   `json:"endpoint"`
	HTTPRetry int      `json:"http_retry"`
	Params    Params   `json:"params"`
}

type TimelineJSON Timeline

type Params struct {
	AppxBlockTimeInMinutes int64 `json:"approx_block_time_in_min"`
	BlocksPerSession       int64 `json:"blocks_per_session"`
}

func (t *Timeline) UnmarshalJSON(data []byte) error {
	tlj := TimelineJSON{}
	err := json.Unmarshal(data, &tlj)
	if err != nil {
		return err
	}
	t.Unit = tlj.Unit
	t.End = int64(math.Abs(float64(tlj.End)) * -1)
	t.Start = int64(math.Abs(float64(tlj.Start)) * -1)
	if t.Start > t.End {
		return NewInvalidStartEndError(t.Start*-1, t.End*-1, t.Unit)
	}
	switch strings.ToLower(t.Unit) {
	case UnitMinutes, UnitMinute, UnitMin, UnitM:
		// all good
	case UnitHours, UnitHour, UnitHr, UnitH:
		// all good
	case UnitDays, UnitDay, UnitD:
		// all good
	case UnitWeeks, UnitWeek, UnitW:
		// all good
	case UnitBlocks, UnitBlock, UnitB:
		// all good
	case UnitSessions, UnitSession, UnitS:
		// all good
	default:
		return NewInvalidUnitError(t.Unit)
	}
	return nil
}

// Gets the conf in the config file
func getConfig(file string) Config {
	fBz, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatalf(err.Error())
	}
	c := Config{}
	err = json.Unmarshal(fBz, &c)
	if err != nil {
		log.Fatalf(err.Error())
	}
	return c
}

func testEndpoint(endpoint string) error {
	r, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return err
	}
	c := http.Client{}
	res, err := c.Do(r)
	if err != nil {
		return err
	}
	if endpoint[len(endpoint)-2:] != "v1" {
		return fmt.Errorf("endpoint must be pocket-core version endpoint")
	}
	if res.StatusCode != 200 {
		return NewHTTPStatusCode(res.StatusCode, "")
	}
	return nil
}

func writeResultFile(result Report, file string) {
	j, err := json.MarshalIndent(result, "", "    ")
	if err != nil {
		log.Fatal(err)
	}
	err = ioutil.WriteFile(file, j, os.ModePerm)
	if err != nil {
		log.Println("ERROR : COULD NOT WRITE REPORT FILE: ", err.Error(), "\nPRINTING TO SCREEN")
		fmt.Println(string(j))
	}
}

func overrideConfig(
	c Config,
	selector string,
	timelineStart int64, timelineEnd int64, timelineUnit string,
	startBlock int64, endBlock int64,
	endpoint string, httpRetry int,
	blocksPerSession int64, blockTimeInMin int64,
) Config {
	log.Println("Processing command line overrides")
	if selector != "" {
		c.Selector = selector
	}

	if timelineStart != -99999 {
		c.Timeline.Start = timelineStart
	}

	if timelineEnd != -99999 {
		c.Timeline.End = timelineEnd
	}

	if timelineUnit != "" {
		c.Timeline.Unit = timelineUnit
	}

	if startBlock != -99999 {
		c.ByBlock.Start = startBlock
	}

	if endBlock != -99999 {
		c.ByBlock.End = endBlock
	}

	if endpoint != "" {
		c.Endpoint = endpoint
	}

	if httpRetry != -1 {
		c.HTTPRetry = httpRetry
	}

	if blocksPerSession != -1 {
		c.Params.BlocksPerSession = blocksPerSession
	}

	if blockTimeInMin != -1 {
		c.Params.AppxBlockTimeInMinutes = blockTimeInMin
	}

	return c
}
