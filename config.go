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
	"time"
)

type Config struct {
	Timeline  Timeline `json:"timeline"`
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
func getConfig() Config {
	file := "config/config.json"
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

func writeResultFile(result Report) {
	j, err := json.MarshalIndent(result, "", "    ")
	if err != nil {
		log.Fatal(err)
	}
	now := time.Now().AddDate(0, 0, -1).Format("01-02-06")
	err = ioutil.WriteFile("result/"+now+".json", j, os.ModePerm)
	if err != nil {
		log.Println("ERROR : COULD NOT WRITE REPORT FILE: ", err.Error(), "\nPRINTING TO SCREEN")
		fmt.Println(string(j))
	}
}
