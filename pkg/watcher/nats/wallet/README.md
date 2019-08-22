## watcher.watch
A request to add wallet to (or remove from) observation:

| Field | Type | Meaning |
| --- | --- | --- |
| 1 wallet | repeated string | Wallet address in Base58 | 
| 2 add | bool | True to add wallet, otherwise to remove it |

### Reply (ACK)
| Field | Type | Meaning |
| --- | --- | --- |
| 1 success | bool | Success is true in case of success | 
| 2 error | string | Error contains error descrition in case of failure |

--- 

## watcher.received
An event from watcher:

| Field | Type | Meaning |
| --- | --- | --- |
| 1 pubkey | string | Wallet address in Base58 |
| 2 token | string | GOLD or MNT |
| 3 amount | string | Token amount in major units: 1.234 (18 decimal places) |
| 4 transaction | string | Digest of the refilling tx in Base58 |

### Reply (ACK)
| Field | Type | Meaning |
| --- | --- | --- |
| 1 success | bool | Success is true in case of success | 
| 2 error | string | Error contains error descrition in case of failure |