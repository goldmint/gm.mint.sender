A set of GM Mint blockchain services.



# Watcher

Watcher service listens Mint blockchain for a new blocks/transactions and detects wallets' incoming transactions. \
ROI (e.g. a set of wallets to observe) could be changed via requests to the service. \
Incoming transactions are saved into storage. Upon saving, the service tries to notify it's consumer.

## Usage

Run the service:
```sh
./watcher \
  --node localhost:4010 \ # Mint node RPC
  --nats localhost:4222 \ # Nats
  --table mywatcher \ # DB table prefix
  --dsn "user:password@tcp(localhost:3306)/mint_sender?collation=utf8_general_ci&timeout=10s&readTimeout=60s&writeTimeout=60s"
```

[Nats messages](https://github.com/void616/gm-mint-sender/tree/master/pkg/watcher/nats/wallet)



# Sender

Sender service handles a queue of payments (outgoing transactions). \
Serves requests to send tokens to the client's wallet from a set of predefined wallets (taking balances into account). \
Upon succesful sending, the service tries to notify it's consumers.

## Usage

Create a file containing private keys of sending wallets (keys.json):
```json
{
	"keys": [
		"PRIVATE KEY 1",
		"PRIVATE KEY 2",
		"PRIVATE KEY N"
	]
}
```

Run the service:
```sh
./sender \
  --keys keys.json \ # Private keys
  --node localhost:4010 \ # Mint node RPC
  --nats localhost:4222 \ # Nats
  --table snd \ # DB table prefix
  --dsn "user:password@tcp(localhost:3306)/mint_sender?collation=utf8_general_ci&timeout=10s&readTimeout=60s&writeTimeout=60s"
```

[Nats messages](https://github.com/void616/gm-mint-sender/tree/master/pkg/sender/nats/sender)



---



## Project

| Dir | Description |
| --- | ----------- |
| build    | Building artifacts, Docker files |
| cmd      | Apps' entrypoint |
| internal | Apps' internals |
| pkg      | Nats messages, Protobuf schemes |
| scripts  | Building scripts |
| vendor   | Go vendoring |



## Transport

Services' primary transport is Nats server. Messages are serialized with Protobuf. \
Nats scheme: `.proto` file contains message scheme, `.go` file contains Nats subject names.



## Storage

Currently there is only MySQL storage implemented, but it's not a problem to replace it with whatever. \
See DAO interfaces for specific service.



## Building

Ensure you have **Go 1.12** to build the project. \
**Protoc** should be installed in order to generate protocol mappings for Nats networking. \
\
Run once to check dependencies: `make deps` \
Build apps:
```sh
make build     # build executables
make dockerize # make a Docker image
```
Makefile builds services for Linux/AMD64 by default. \
Use `TARGETS` arg to change this behaviour. For instance:
```sh
make TARGETS="watcher/windows/amd64/ sender/windows/amd64/"
```



## Testing

Setup MySQL database, run Nats, run Mint services and then run `go run cmd/testcli/main.go` to interact with the services.

### Code test
Nats transport tests are require running Nats server (localhost:4222 by default). \
MySQL DAO tests are require running MySQL Server ~5.7. DSN string should be manually edited. \
\
Packages unit tests:
```sh
make test
```



## TODO

- [x] Instrumenting (Prometheus)
- [ ] HTTP API
