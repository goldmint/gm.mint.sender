## mintsender.sender.send
A request to send a token to an address.
```
message {
  string service = 1;    // Service name (to differentiate multiple requestors): 1..64
  string id = 2;         // Unique request ID (within service): 1..64
  string publicKey = 3;  // Destination wallet address in Base58
  string token = 4;      // GOLD or MNT
  string amount = 5;     // Token amount in major units: 1.234 (18 decimal places)
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

## mintsender.sender.sent
An event with sending completion status.
```
message {
  bool success = 1;        // Success is true in case of success
  string error = 2;        // Error contains error descrition in case of failure
  string service = 3;      // Service name (to differentiate multiple requestors): 1..64
  string id = 4;           // Unique request ID: 1..64
  string publicKey = 5;    // Destination wallet address in Base58 (empty on failure)
  string token = 6;        // GOLD or MNT (empty on failure)
  string amount = 7;       // Token amount in major units: 1.234 (18 decimal places, empty on failure)
  string transaction = 8;  // Transaction digest in Base58 (empty on failure)
}
```
### Reply (ACK)
```
message {
  bool success = 1;  // Success is true in case of success
  string error = 2;  // Error contains error descrition in case of failure
}