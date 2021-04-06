# relay_counter
A simple tool that counts relays during a time period

### How To Use

1) Ensure golang 1.13+ environment is properly installed
2) Sync (RC-0.5.2.9) Pocket Core Instance (ensure completely synced)
3) go get github.com/pokt-network/relay_counter
4) cd <GOPATH>/src/github.com/pokt-network/relay_counter
5) Edit configuration file (See details below)
6) go run ./...
7) check <GOPATH>/src/github.com/pokt-network/relay_counter/result for the result file

### Config File
Found in <path to relay_counter>/config/config.json
```
{
  "timeline": {
    "start": -4,
    "end": -1, // can be 0 for latest
    "unit": "days" // blocks, sessions, min, weeks, hours
  },
  "endpoint": "http://localhost:8081/v1",
  "http_retry": 3, // if rpc not responsive
  "params": {
    "blocks_per_session": 4, // needed for sessions
    "approx_block_time_in_min": 15
  }
}
```
  
### TL;DR how it works
With a simple config file, relay_counter uses binary search to find the nearest blocks to the start/end, then tallies up the relays using valid claims and proofs transactions. 

You can pass arguments like a different `config` file path or a `results` file path. Also you can override on the fly with arguments any of the parameters that exists into the config.json.

```bash
go run ./... -config=test/config.json -results test/result.json -endpoint=https://some.node.com/v1
```

In any case you can use `-h` to see all the available options.

#### Config.json | CLI args
| Config File Option             | CLI Arg           | Description                                                 | Options/Default                                      |
|--------------------------------|-------------------|-------------------------------------------------------------|------------------------------------------------------|
| -                              | -config           | config file path                                            | config/config.json                                   |
| -                              | -results          | results file path                                           | result/<date>.json                                   |
| selector                       | -selector         | Use this to point which method will you use to select block | timeline, byClock                                    |
| timeline.start                 | -timelineStart    | used only when selector=timeline                            |                                                      |
| timeline.end                   | -timelineEnd      | used only when selector=timeline                            |                                                      |
| timeline.unit                  | -timelineUnit     | used only when selector=timeline                            | block[s],session[s],minute[s],hour[s],day[s],week[s] |
| byBlock.start                  | -startBlock       | used only when selector=byBlock                             |                                                      |
| byBlock.end                    | -endBlock         | used only when selector=byBlock                             |                                                      |
| endpoint                       | -endpoint         | endpoint must be pocket-core version endpoint               |                                                      |
| http_retry                     | -httpRetry        | how much retries will be done in case some endpoint fail    |                                                      |
| params.block_per_session       | -blocksPerSession |                                                             |                                                      |
| parms.approx_block_time_in_min | -blockTimeInMin   | approximate time before next block height been generated    |                                                      |