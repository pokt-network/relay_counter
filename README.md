# relay_counter
A simple tool that counts relays during a time period

### How To Use

1) Ensure golang 1.13+ environment is properly installed
2) Sync Pocket Core Instance (ensure completely synced)
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
