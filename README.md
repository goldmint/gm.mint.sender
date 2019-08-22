A set of services to integrate an accounting with Mint blockchain.



# Watcher
Service listens Mint blockchain for new blocks/transactions and detects wallets refilling (incoming transactions). \
ROI (e.g. a set of wallets to observe) could be changed via requests to the service. \
Refilling transactions are saved into storage. \
Upon saving, the service tries to notify it's consumers about a refilling.
## Usage
Run the service:
```sh
./watcher \
  --node localhost:4010 \ # Mint node RPC
  --nats localhost:4222 \ # Nats
  --table wtc \ # DB table prefix
  --dsn "user:password@tcp(localhost:3306)/watcher?collation=utf8_general_ci&timeout=10s&readTimeout=60s&writeTimeout=60s"
```
[Nats messages](https://github.com/void616/gm-mint-sender/tree/master/pkg/watcher/nats/wallet)



# Sender
Service handles a queue of payments (outgoing transactions). \
Serves requests to send tokens to the client's wallet from a set of predefined wallets (taking balances into account). \
Upon succesful sending, the service tries to notify it's consumers about success.
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
  --dsn "user:password@tcp(localhost:3306)/sender?collation=utf8_general_ci&timeout=10s&readTimeout=60s&writeTimeout=60s"
```
[Nats messages](https://github.com/void616/gm-mint-sender/tree/master/pkg/sender/nats/sender)



## Transport
Services primary transport is Nats server with Protobuf serialization. \
`.proto` file contains message scheme, `.go` file contains Nats subject names. \
For instance, [here](https://github.com/void616/gm-mint-sender/tree/master/pkg/watcher/nats/wallet) or [here](https://github.com/void616/gm-mint-sender/tree/master/pkg/sender/nats/sender).



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
