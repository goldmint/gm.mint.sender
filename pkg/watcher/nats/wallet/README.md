## mintsender.watcher.watch
A request to add wallet to (or remove from) observation.
```
message {
  string service = 1;             // Service name (to differentiate multiple requestors): 1..64
  repeated string publicKey = 2;  // Wallet address in Base58
  bool add = 3;                   // True to add wallet, otherwise to remove it
}
```
### Reply (ACK)
```
message {
  bool success = 1;  // Success is true in case of success
  string error = 2;  // Error contains error descrition in case of failure
}
```

---

## mintsender.watcher.refill
An event from watcher.
```
message RefillEvent {
  string service = 1;      // Service name (to differentiate multiple requestors): 1..64
	string publicKey = 2;    // Destination (watching) wallet address in Base58
	string from = 3;         // Source wallet address in Base58
	string token = 4;        // GOLD or MNT
	string amount = 5;       // Token amount in major units: 1.234 (18 decimal places)
	string transaction = 6;  // Digest of the refilling tx in Base58
}
```
### Reply (ACK)
```
message {
  bool success = 1;  // Success is true in case of success
  string error = 2;  // Error contains error descrition in case of failure
}
```