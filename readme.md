# Coinbase VWAP

Connects to the Coinbase API, subscribes to the matches channel and pulls data for
BTC-USD, ETH-USD & ETH-BTC. A VWAP is output to the console for each of these as
match messages are read.

## Requires
* Dependencies outlined in [./build/docker/Dockerfile](./build/docker/Dockerfile).
* Docker (tested with 20.10.12).

## Setup
N/A

## Development
```
make help
```

Make targets can be invoked on your host machine as normal or in a container. See
[./build/docker/README.md](./build/docker/README.md).

## Usage

Run the app:

```
make run
```

## Layout
    .
    ├── cmd                     
    │   └── coinbasevwap        # Application entrypoint.
    └── build                   # Scripts used in the build pipeline.

## Design / Assumptions / Misc

### Coinbase API & Domain

I made a few assumptions here:
- Match price is price per unit
- Match size is the number of units
- Matches missing (gaps indicated by non-consecutive sequence values) weren't important
for this exercise.
- We wanted to calculate the VWAP for both buy/sell Matches, especially given the 
above.

### Use of floating point for currency calculations

Ideally, something like https://github.com/shopspring/decimal may have been more 
suitable to avoid floating point non-deterministic calculations.

### Configuration

There are some values hardcoded that may have better been derived from configuration. 
For example, the Coinbase feed URL.

### Auxiliary stuff re-used

There's a few (non-source file) bits and pieces I re-used from my other personal
projects (e.g shell scripts, docker files, makefile)

### Graceful shutdown was difficult

Interrupting the application will see graceful shutdown of at least ETH-BTC fail. 
Each connection is allowed 3 seconds to close but the websocket read blocks. ETH-BTC
in particular seemed to go more than a minute between reads.

### Logging

I skipped any proper logging implementation here, just opting for use of the standard
package.

### Integration testing wasn't completely black-box

This would of been a bit difficult to achieve and, in the interest of time, I kept
it brief. See [here](./cmd/coinbasevwap/main_test.go).