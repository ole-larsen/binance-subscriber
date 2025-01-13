# Binance Integration

This project is a simple websocket application for interacting with the Binance. 
## Technology stack

- golang

## build
```
go build -o binance cmd/server/main.go
```
## run
```
./binance -i=-i=btcusdt@depth,ethusdt@depth -a=localhost:8080 
```
## test

run application then run index.html