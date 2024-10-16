# Trading 212 Parser

This parser should parse the buy/sell transactions from history files exported from Trading 212 to output the yearly Irish CGT liability (best effort/close enough basis) using the FIFO method to calculate profits.

**NOTE:** I take no responsibilty if this does not calculate your taxes correctly. Always consult a tax advisor for legal matters. 

## Running

* Populate the `./configs/config.json` file with a map of `Year` to `Path` of the history file exported from Trading 212
* Run `go run cmd/main.go -config configs/config.json`
* Run `go run cmd/main.go --help` for usage

The output text will show you your expected tax liability for all the years.
