GM Mint blockchain services.

---

# Sender

Sender service handles a queue of token transferring transactions. \
Serves requests to send tokens to the client's wallet from a set of predefined wallets (taking balances into account). \
Is able to set 'approved' tag. In this case 'authority' wallet should be defined. \
Upon succesful sending, the service tries to notify it's consumers.

## Usage

Prepare `config.yaml`:
```yaml
# Log
log:
  level: debug
  color: yes
  json: no
# API (at least one)
api:
  # Nats interface 
  nats:
    url: 127.0.0.1:4222
    prefix: "mint"
  # HTTP interface
  http:
    port: 8080
# Database
db:
  driver: mysql
  dsn: user:password@tcp(127.0.0.1:3306)/database?collation=utf8_general_ci&timeout=10s&readTimeout=60s&writeTimeout=60s
  prefix: sender
# Prometheus metrics (optional)
metrics: 0
# GM Alerts (optional)
gcloud_alerts: false
# Mint nodes (at least one)
nodes:
  - 127.0.0.1:4010
# Service wallets
wallets:
  - PRIVATE_KEY
  - PRIVATE_KEY
```

Run the service:
```sh
./sender
```

[Nats messages](pkg/sender/nats)
[HTTP messages](pkg/sender/http/README.md)

---

# Watcher

Watcher service listens Mint blockchain for a new blocks/transactions and detects wallets' incoming transactions. \
ROI (e.g. a set of wallets to observe) could be changed via requests to the service. \
Incoming transactions are saved into storage. Upon saving, the service tries to notify it's consumer.

## Usage

Prepare `config.yaml`:
```yaml
# Log
log:
  level: debug
  color: yes
  json: no
# API (at least one)
api:
  # Nats interface
  nats:
    url: 127.0.0.1:4222
    prefix: "mint"
  # HTTP interface
  http:
    port: 9001
# Database
db:
  driver: mysql
  dsn: user:password@tcp(127.0.0.1:3306)/database?collation=utf8_general_ci&timeout=10s&readTimeout=60s&writeTimeout=60s
  prefix: watcher
# Prometheus metrics (optional)
metrics: 0
# GM Alerts (optional)
gcloud_alerts: false
# Mint nodes (at least one)
nodes:
  - 127.0.0.1:4010

```

Run the service:
```sh
./watcher
```

[Nats messages](pkg/watcher/nats)
[HTTP messages](pkg/watcher/http/README.md)

---

## Project

| Dir | Description |
| --- | ----------- |
| build    | Built artifacts, Docker files |
| cmd      | Apps' entrypoint |
| internal | Apps' internals |
| pkg      | Nats/HTTP messages, Protobuf schemes |
| scripts  | Building scripts |
| vendor   | Go vendoring |

## Transport

Services' primary transport is Nats server. Messages are serialized with Protobuf. \
Nats scheme: `.proto` file contains message scheme, `.go` file contains Nats subject names. \
Simple HTTP API is supported as well.

## Storage

Currently there is only MySQL storage implemented, but it's not a problem to replace it with whatever. \
See DAO interfaces for specific service.

## Building

Ensure you have at least **Go 1.12** to build the project. \
**Protoc** should be installed in order to generate protocol mappings for Nats networking. \
\
Run once to check dependencies: `make deps` \
Build apps:
```sh
make build  # build executables
make image  # make Docker image
```
Makefile builds services for Linux/AMD64 by default. \
Use `TARGETS` arg to change this behaviour. For instance:
```sh
make TARGETS="watcher/windows/amd64/ sender/windows/amd64/ testcli/windows/amd64"
```

## Testing

Setup MySQL database, run Nats, run Mint services and then run `go run cmd/cli/main.go` to interact with the services. \
![Test Cli](docs/testcli.png)

### Code test
Packages unit tests:
```sh
make test
```
