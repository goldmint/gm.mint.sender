## sender.send
A request to send a token to an address.

| Field | Type | Meaning |
| --- | --- | --- |
| 1 id | string | Unique request ID: 1..64 |
| 2 pubkey | string | Destination wallet address in Base58 |
| 3 token | string | GOLD or MNT |
| 4 amount | string | Token amount in major units: 1.234 (18 decimal places) |

### Reply (ACK)
| Field | Type | Meaning |
| --- | --- | --- |
| 1 success | bool | Success is true in case of success | 
| 2 error | string | Error contains error descrition in case of failure |

---

## sender.sent
An event with sending completion status.

| Field | Type | Meaning |
| --- | --- | --- |
| 1 id | string | Unique request ID: 1..64 |
| 2 success | bool | Success is true in case of success |
| 3 error | string | Error contains error descrition in case of failure |
| 4 transaction | string | Digest of the refilling tx in Base58 on success or an empty string on failure |

### Reply (ACK)
| Field | Type | Meaning |
| --- | --- | --- |
| 1 success | bool | Success is true in case of success | 
| 2 error | string | Error contains error descrition in case of failure |