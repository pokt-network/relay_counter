package main

import (
	"bytes"
	"encoding/json"
	"github.com/pokt-network/pocket-core/app/cmd/rpc"
	"github.com/pokt-network/pocket-core/codec"
	types3 "github.com/pokt-network/pocket-core/codec/types"
	"github.com/pokt-network/pocket-core/crypto"
	pc "github.com/pokt-network/pocket-core/types"
	appTypes "github.com/pokt-network/pocket-core/x/apps/types"
	"github.com/pokt-network/pocket-core/x/auth"
	authTypes "github.com/pokt-network/pocket-core/x/auth"
	"github.com/pokt-network/pocket-core/x/auth/types"
	govTypes "github.com/pokt-network/pocket-core/x/gov"
	nodeTypes "github.com/pokt-network/pocket-core/x/nodes/types"
	pcTypes "github.com/pokt-network/pocket-core/x/pocketcore/types"
	cryptoamino "github.com/tendermint/tendermint/crypto/encoding/amino"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
)

//type Block coretypes.ResultBlock

const (
	BlockTxsPath = "/query/blocktxs"
	StatePath    = "/query/state"
	HeightPath   = "/query/height"
	BlockPath    = "/query/block"
	SupplyPath   = "/query/supply"
	UnitBlocks   = "blocks"
	UnitBlock    = "block"
	UnitB        = "b"
	UnitSessions = "sessions"
	UnitSession  = "session"
	UnitS        = "s"
	UnitMinutes  = "minutes"
	UnitMinute   = "minute"
	UnitMin      = "min"
	UnitM        = "m"
	UnitHours    = "hours"
	UnitHour     = "hour"
	UnitHr       = "hr"
	UnitH        = "h"
	UnitDays     = "days"
	UnitDay      = "day"
	UnitD        = "d"
	UnitWeeks    = "weeks"
	UnitWeek     = "week"
	UnitW        = "w"
)

var (
	cdc = codec.NewCodec(types3.NewInterfaceRegistry())
)

func init() {
	cdc.SetUpgradeOverride(false)
	pc.RegisterCodec(cdc)
	pcTypes.RegisterCodec(cdc)
	authTypes.RegisterCodec(cdc)
	nodeTypes.RegisterCodec(cdc)
	appTypes.RegisterCodec(cdc)
	govTypes.RegisterCodec(cdc)
	crypto.RegisterAmino(cdc.AminoCodec().Amino)
	cryptoamino.RegisterAmino(cdc.AminoCodec().Amino)
	codec.RegisterEvidences(cdc.AminoCodec(), cdc.ProtoCodec())
}

type Timeline struct {
	Start int64  `json:"start"`
	End   int64  `json:"end"`
	Unit  string `json:"unit"`
}

type ByBlock struct {
	Start int64 `json:"start"`
	End   int64 `json:"end"`
}

type PaginatedHeightParams struct {
	Height  int64  `json:"height"`
	Page    int    `json:"page,omitempty"`
	PerPage int    `json:"per_page,omitempty"`
	Prove   bool   `json:"prove,omitempty"`
	Sort    string `json:"order,omitempty"`
}

type AppState struct {
	PocketCoreState `json:"pocketcore"`
}

type PocketCoreState struct {
	Claims []pcTypes.MsgClaim `json:"claims"`
}

type StateRPCResponse struct {
	AppState AppState `json:"app_state"`
}

type HeightRPCResponse struct {
	Height int64 `json:"height"`
}

type SupplyRPCResponse struct {
	Total string `json:"total"`
}

type TimelineReport struct {
	MinHeight int64 `json:"min_height"`
	MaxHeight int64 `json:"max_height"`
}

type Report struct {
	TotalRelaysCompleted     int64                 `json:"total_relays_completed"`
	TotalChallengesCompleted int64                 `json:"total_challenges_completed"`
	TotalMinted              int64                 `json:"total_minted"`
	TotalGoodTxs             int64                 `json:"total_good_txs"`
	TotalBadTxs              int64                 `json:"total_bad_txs"`
	TotalProofTxs            int64                 `json:"proof_msgs"`
	BadTxsMap                map[uint32]int64      `json:"bad_txs_count_by_error"`
	NodeReports              map[string]NodeReport `json:"node_report"`
	AppReports               map[string]AppReport  `json:"app_report"`
	BlockSelector            string                `json:"selector"`
	TimelineReport           TimelineReport        `json:"timeline_report"`
	ByBlockReport            ByBlock               `json:"by_block"`
}

type ServiceReport struct {
	Address     string `json:"address"`
	TotalRelays int64  `json:"total_relays"`
	ChainID     string `json:"relay_chain"`
}

type NodeReport struct {
	Service              []ServiceReport  `json:"serviced"`
	TotalRelays          int64            `json:"total_relays"`
	ServiceReportByChain map[string]int64 `json:"service_by_chain"`
}

type AppReport struct {
	ServicedBy            []ServiceReport  `json:"serviced_by"`
	TotalRelays           int64            `json:"total_relays"`
	ServicedReportByChain map[string]int64 `json:"serviced_by_chain"`
}

type ClaimsMap map[int64][]pcTypes.MsgClaim
type BlockTxsMap map[int64]rpc.RPCResultTxSearch

func ConvertTimelineToHeights(config Config) (timelineReport TimelineReport, err error) {
	// start and end are negative values
	var startInBlocks, endInBlocks, minHeight, maxHeight int64
	var targetStartTime, targetEndTime time.Time
	log.Println("Getting the latest height")
	// get the latest height
	latestheight, err := GetLatestHeight(config)
	if err != nil {
		return timelineReport, err
	}
	log.Println("Getting the latest block")
	block, err := GetBlock(latestheight, config)
	if err != nil {
		return timelineReport, err
	}
	latestHeight := block.Block.Height
	latestTime := block.Block.Time
	log.Printf("Latest height is %d and latest time is: %s\n", latestHeight, latestTime.String())
	switch strings.ToLower(config.Timeline.Unit) {
	case UnitMinutes, UnitMinute, UnitMin, UnitM:
		log.Println("Timeline unit is minutes")
		targetStartTime, targetEndTime = GetTargetTimes(config, latestTime, time.Minute)
		minHeight, maxHeight = GetClosestHeights(latestHeight, targetStartTime, latestTime, targetEndTime, config)
	case UnitHours, UnitHour, UnitHr, UnitH:
		log.Println("Timeline unit is hours")
		targetStartTime, targetEndTime = GetTargetTimes(config, latestTime, time.Hour)
		minHeight, maxHeight = GetClosestHeights(latestHeight, targetStartTime, latestTime, targetEndTime, config)
	case UnitDays, UnitDay, UnitD:
		log.Println("Timeline unit is days")
		targetStartTime, targetEndTime = GetTargetTimes(config, latestTime, time.Hour*24)
		minHeight, maxHeight = GetClosestHeights(latestHeight, targetStartTime, latestTime, targetEndTime, config)
	case UnitWeeks, UnitWeek, UnitW:
		log.Println("Timeline unit is weeks")
		targetStartTime, targetEndTime = GetTargetTimes(config, latestTime, time.Hour*24*7)
		minHeight, maxHeight = GetClosestHeights(latestHeight, targetStartTime, latestTime, targetEndTime, config)
	case UnitBlocks, UnitBlock, UnitB:
		log.Println("Timeline unit is blocks")
		minHeight = latestHeight + config.Timeline.Start
		maxHeight = latestHeight + config.Timeline.End
	case UnitSessions, UnitSession, UnitS:
		log.Println("Timeline unit is sessions")
		startInBlocks = config.Timeline.Start * config.Params.BlocksPerSession
		endInBlocks = config.Timeline.End * config.Params.BlocksPerSession
		minHeight = latestHeight + startInBlocks
		maxHeight = latestHeight + endInBlocks
	default:
		panic("ERROR: unrecognized unit: (minutes, hours, days, weeks, blocks)")
	}
	if minHeight < 0 {
		err = NewInvalidMinimumHeightError(minHeight)
		return
	}
	timelineReport = TimelineReport{
		MinHeight: minHeight,
		MaxHeight: maxHeight,
	}
	return
}

func GetChainData(minHeight, maxHeight int64, config Config) (blockTxsMap BlockTxsMap, claimsMap ClaimsMap, supplyStart, supplyEnd int) {
	log.Println("Beginning Chain Data Operations")
	count := 0
	blockTxsMap = make(BlockTxsMap, 0)
	claimsMap = make(ClaimsMap, 0)
	// loop through all the heights and retrieve all the block-txs
	log.Printf("Begin transactions / claims retrieval for heights: %d through %d\n", minHeight, maxHeight)
	for height := minHeight; height < maxHeight; height++ {
		if _, ok := blockTxsMap[height]; !ok {
			result, err := GetBlockTx(height, config)
			if err != nil {
				if count >= config.HTTPRetry {
					log.Fatalf("After %d retries, unable to get block-txs for height: %d with error: %s", config.HTTPRetry, height, err.Error())
				} else {
					log.Printf("RPC failure for blocktxs: %s. Trying to retry. Retry count is: %d/%d\n", err.Error(), count, config.HTTPRetry)
					count++
					height-- // try the same height again
					// arbitrary sleep to retry
					time.Sleep(1 * time.Second)
					continue
				}
			} else {
				log.Printf("BlkTxs retrieved for height: %d, %d out of %d\n", height, height-minHeight, maxHeight-minHeight)
				// add the block-txs to the result
				blockTxsMap[height] = result
				count = 0
			}
		}
		// skip claims for blocks 0 and 1
		if height == 0 || height == 1 {
			continue
		}
		// we want to check the claim at height - 1 cause the state = endBlockState
		if _, ok := claimsMap[height]; !ok {
			// get the claim for height-1
			claimsResult, err := GetClaims(height-1, config)
			if err != nil {
				if count >= config.HTTPRetry {
					log.Fatalf("After %d retries, unable to get claims for height: %d, with error: %s", config.HTTPRetry, height, err.Error())
				} else {
					log.Printf("RPC failure for claims: %s\nTrying to retry. Retry count is: %d/%d\n", err.Error(), count, config.HTTPRetry)
					count++
					height-- // try the same height again
					// arbitrary sleep to retry
					time.Sleep(5 * time.Second)
					continue
				}
			} else {
				log.Printf("Claims retrieved for height: %d, %d out of %d\n", height, height-minHeight, maxHeight-minHeight)
				// add the block-txs to the result
				claimsMap[height] = claimsResult
				count = 0
			}
		}
	}
	log.Println("Getting starting supply")
	// get the beginning and end supply
	supplyStart, err := GetSupply(minHeight, config)
	if err != nil {
		log.Fatalf("unable to get the supply at height: %d with error %s", minHeight, err.Error())
	}
	log.Println("Getting ending supply")
	supplyEnd, err = GetSupply(maxHeight, config)
	if err != nil {
		log.Fatalf("unable to get the supply at height: %d with error %s", maxHeight, err.Error())
	}
	return
}

func ProcessChainData(txsMap BlockTxsMap, claimsMap ClaimsMap, supplyStart, supplyEnd int, selector string, timelineReport TimelineReport, byBlock ByBlock) (result Report) {
	log.Println("Chain Data Process Operation Started")
	result = Report{
		BadTxsMap:      make(map[uint32]int64),
		NodeReports:    make(map[string]NodeReport, 0),
		AppReports:     make(map[string]AppReport, 0),
		BlockSelector:  selector,
		TimelineReport: timelineReport,
		ByBlockReport:  byBlock,
	}
	log.Println("Looping through all of the block-txs and matching them with the corresponding claims")
	for height, blockTx := range txsMap {
		for _, txResult := range blockTx.Txs {
			// check if bad transaction
			if txResult.TxResult.Code != 0 {
				log.Println("Bad tx found and logged")
				result.TotalBadTxs++
				result.BadTxsMap[txResult.TxResult.Code]++
				continue
			}
			// log good tx
			result.TotalGoodTxs++
			// if not proofTx, continue on
			if txResult.StdTx.Msg.Type() != pcTypes.MsgProofName {
				log.Println("Good non-proof tx found and logged")
				continue
			}
			// this is a proof msg
			proofMsg, ok := txResult.StdTx.Msg.(pcTypes.MsgProof)
			if !ok {
				log.Fatalf(NewProofMsgInterfaceError().Error())
			}
			// log good tx
			result.TotalProofTxs++
			log.Println("Proof tx found and logged")
			claim := pcTypes.MsgClaim{}
			// find the corresponding claim
			for _, c := range claimsMap[height] {
				if !c.FromAddress.Equals(proofMsg.GetSigner()) {
					continue
				}
				claim = c
			}
			if claim.FromAddress == nil {
				log.Fatalf("No claim for valid proof object...")
			}
			log.Println("Corresponding claim found")
			// check to see if claim is for relays
			et := claim.EvidenceType
			if et != pcTypes.RelayEvidence {
				result.TotalChallengesCompleted++
				continue
			}
			// get appAddress
			appAddress := GetAddressFromPubKey(claim.SessionHeader.ApplicationPubKey)
			nodeAddress := claim.FromAddress.String()
			// get total # of relays
			totalRelays := claim.TotalProofs
			// get the relay chain id
			chainID := claim.SessionHeader.Chain
			// retrieve the app/node reports
			appReport, found := result.AppReports[appAddress]
			if !found {
				log.Printf("New App report created for address %s\n", appAddress)
				appReport = NewAppReport()
			}
			nodeReport, found := result.NodeReports[nodeAddress]
			if !found {
				log.Printf("New Node report created for address %s\n", nodeAddress)
				nodeReport = NewNodeReport()
			}
			log.Printf("Adding data to the node report for address:%s\n", nodeAddress)
			log.Printf("Adding data to the app report for address:%s\n", appAddress)
			// add to the reports totals
			appReport.TotalRelays += totalRelays
			nodeReport.TotalRelays += totalRelays
			// add to the chain statistics
			appReport.ServicedReportByChain[chainID] += totalRelays
			nodeReport.ServiceReportByChain[chainID] += totalRelays
			result.TotalRelaysCompleted += totalRelays
			// add an individual service report to the appReport
			appReport.ServicedBy = append(appReport.ServicedBy, ServiceReport{
				Address:     nodeAddress,
				TotalRelays: totalRelays,
				ChainID:     chainID,
			})
			// add an individual service report to the nodeReport
			nodeReport.Service = append(nodeReport.Service, ServiceReport{
				Address:     appAddress,
				TotalRelays: totalRelays,
				ChainID:     chainID,
			})
			// set the reports in the master report
			result.AppReports[appAddress] = appReport
			result.NodeReports[nodeAddress] = nodeReport
		}
	}
	log.Println("Calculating the total minted")
	// set the supply difference as total minted
	result.TotalMinted = int64(supplyEnd - supplyStart)
	log.Println("Report created")
	return result
}

func NewAppReport() AppReport {
	return AppReport{
		ServicedBy:            make([]ServiceReport, 0),
		TotalRelays:           0,
		ServicedReportByChain: make(map[string]int64),
	}
}

func NewNodeReport() NodeReport {
	return NodeReport{
		Service:              make([]ServiceReport, 0),
		TotalRelays:          0,
		ServiceReportByChain: make(map[string]int64),
	}
}

func GetAddressFromPubKey(pkHex string) string {
	apk, err := crypto.NewPublicKey(pkHex)
	if err != nil {
		log.Fatalf(NewPublicKeyError().Error())
	}
	return apk.Address().String()
}

func GetTargetTimes(config Config, latestTime time.Time, unit time.Duration) (targetStartTime, targetEndTime time.Time) {
	log.Println("Calclating the approximate start and end times")
	st := time.Duration(config.Timeline.Start) * unit
	et := time.Duration(config.Timeline.End) * unit
	targetStartTime = latestTime.Add(st)
	targetEndTime = latestTime.Add(et)
	log.Printf("Target start: %s\nTarget End: %s\n", targetStartTime.String(), targetEndTime.String())
	return
}

func GetClosestHeights(latestHeight int64, targetStartTime, latestBlockTime, targetEndTime time.Time, config Config) (startHeight, endHeight int64) {
	log.Println("Begin Closest Height Operations")
	appxStartHeight := latestHeight - int64(latestBlockTime.Sub(targetStartTime).Minutes()/15)
	appxEndHeight := latestHeight - int64(latestBlockTime.Sub(targetEndTime).Minutes()/15)
	startHeight = BlockBinarySearch(targetStartTime, latestHeight, appxStartHeight, config)
	log.Printf("Closest Start Height Found: %d\n", startHeight)
	endHeight = BlockBinarySearch(targetEndTime, latestHeight, appxEndHeight, config)
	log.Printf("Closest End Height Found: %d\n", endHeight)
	return
}

func BlockBinarySearch(targetStartTime time.Time, latestHeight, tryHeight int64, config Config) (closestHeight int64) {
	log.Printf("Performing a binary search for the closest height to the target time: %s\n", targetStartTime.String())
	max := latestHeight
	closestHeight = tryHeight
	closestTime := time.Time{}
	httpTryCount := 0
	for min := int64(0); min < max && max-min != 1; {
		log.Println("min: ", min, "max", max, "try height", tryHeight, "closest height", closestHeight)
		// get the latest height
		block, err := GetBlock(tryHeight, config)
		// retry logic
		if err != nil {
			if httpTryCount >= config.HTTPRetry {
				log.Fatalf("After %d retries, unable to get block for height: %d, with error: %s", config.HTTPRetry, tryHeight, err.Error())
			} else {
				log.Printf("RPC failure for block by height: %s\nTrying to retry. Retry count is: %d/%d\n", err.Error(), httpTryCount, config.HTTPRetry)
				httpTryCount++
				// arbitrary sleep to retry
				time.Sleep(5 * time.Second)
				continue
			}
		}
		// if tryHeight block is before our target...
		if block.Block.Time.Before(targetStartTime) {
			// minimum is where the pivot was
			min = tryHeight
		} else {
			// maximum is where the pivot was
			max = tryHeight
		}
		// see if target height is closer than the current closest height
		if IsCloserThan(block.Block.Time, closestTime, targetStartTime) {
			// if is closer, let's update the closest
			closestTime = block.Block.Time
			closestHeight = block.Block.Height
		}
		// new pivot
		tryHeight = (min + max) / 2
		// reset the http try
		httpTryCount = 0
	}
	return
}

func IsCloserThan(check, other, target time.Time) bool {
	diff1 := math.Abs(float64(target.Sub(check).Nanoseconds()))
	diff2 := math.Abs(float64(target.Sub(other).Nanoseconds()))
	if diff1 >= diff2 {
		return false
	}
	return true
}

func GetBlockTx(height int64, config Config) (result rpc.RPCResultTxSearch, err error) {
	requestBody := PaginatedHeightParams{
		Height:  height,
		PerPage: 10000000,
	}
	r, err := json.Marshal(requestBody)
	if err != nil {
		return result, err
	}
	req, err := http.NewRequest("POST", config.Endpoint+BlockTxsPath, bytes.NewBuffer(r))
	if err != nil {
		return result, err
	}
	c := http.Client{}
	res, err := c.Do(req)
	if err != nil {
		return result, err
	}
	defer res.Body.Close()
	bodyBz, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return result, err
	}
	if res.StatusCode != 200 {
		return result, NewHTTPStatusCode(res.StatusCode, string(bodyBz))
	}
	rts := &coretypes.ResultTxSearch{}
	err = json.Unmarshal(bodyBz, &rts)
	if err != nil {
		return result, err
	}
	result = ResultTxSearchToRPC(rts)
	return result, err
}

func GetClaims(height int64, config Config) (result []pcTypes.MsgClaim, err error) {
	requestBody := PaginatedHeightParams{
		Height:  height,
		PerPage: 10000000,
	}
	r, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", config.Endpoint+StatePath, bytes.NewBuffer(r))
	if err != nil {
		return nil, err
	}
	c := http.Client{}
	res, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	bodyBz, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != 200 {
		return nil, NewHTTPStatusCode(res.StatusCode, string(bodyBz))
	}
	state := StateRPCResponse{}
	err = cdc.UnmarshalJSON(bodyBz, &state)
	s := string(bodyBz)
	s = s
	return state.AppState.Claims, err
}

func GetLatestHeight(config Config) (int64, error) {
	req, err := http.NewRequest("POST", config.Endpoint+HeightPath, nil)
	if err != nil {
		return 0, err
	}
	c := http.Client{}
	res, err := c.Do(req)
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()
	bodyBz, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return 0, err
	}
	if res.StatusCode != 200 {
		return 0, NewHTTPStatusCode(res.StatusCode, string(bodyBz))
	}
	height := HeightRPCResponse{}
	err = json.Unmarshal(bodyBz, &height)
	return height.Height, err
}

func GetBlock(height int64, config Config) (block *coretypes.ResultBlock, err error) {
	requestBody := PaginatedHeightParams{Height: height}
	r, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", config.Endpoint+BlockPath, bytes.NewBuffer(r))
	if err != nil {
		return nil, err
	}
	c := http.Client{}
	res, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	bodyBz, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != 200 {
		return nil, NewHTTPStatusCode(res.StatusCode, string(bodyBz))
	}
	err = cdc.UnmarshalJSON(bodyBz, &block)
	return
}

func GetSupply(height int64, config Config) (supply int, err error) {
	requestBody := PaginatedHeightParams{Height: height}
	r, err := json.Marshal(requestBody)
	if err != nil {
		return 0, err
	}
	req, err := http.NewRequest("POST", config.Endpoint+SupplyPath, bytes.NewBuffer(r))
	if err != nil {
		return 0, err
	}
	c := http.Client{}
	res, err := c.Do(req)
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()
	bodyBz, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return 0, err
	}
	if res.StatusCode != 200 {
		return 0, NewHTTPStatusCode(res.StatusCode, string(bodyBz))
	}
	s := SupplyRPCResponse{}
	err = cdc.UnmarshalJSON(bodyBz, &s)
	if err != nil {
		return 0, err
	}
	supply, err = strconv.Atoi(s.Total)
	return
}

func ResultTxSearchToRPC(res *coretypes.ResultTxSearch) rpc.RPCResultTxSearch {
	if res == nil {
		return rpc.RPCResultTxSearch{}
	}
	rpcTxSearch := rpc.RPCResultTxSearch{
		Txs:        make([]*rpc.RPCResultTx, 0, res.TotalCount),
		TotalCount: res.TotalCount,
	}
	for _, result := range res.Txs {
		rpcTxSearch.Txs = append(rpcTxSearch.Txs, ResultTxToRPC(result))
	}
	return rpcTxSearch
}

func ResultTxToRPC(res *coretypes.ResultTx) *rpc.RPCResultTx {
	if res == nil {
		return nil
	}
	tx := UnmarshalTx(res.Tx, res.Height)
	r := &rpc.RPCResultTx{
		Hash:     res.Hash,
		Height:   res.Height,
		Index:    res.Index,
		TxResult: res.TxResult,
		Tx:       res.Tx,
		Proof:    res.Proof,
		StdTx:    tx,
	}
	return r
}

func UnmarshalTx(txBytes []byte, height int64) types.StdTx {
	defaultTxDecoder := auth.DefaultTxDecoder(cdc)
	tx, err := defaultTxDecoder(txBytes, height)
	if err != nil {
		log.Fatalf("Could not decode transaction: " + err.Error())
	}
	return tx.(auth.StdTx)
}
