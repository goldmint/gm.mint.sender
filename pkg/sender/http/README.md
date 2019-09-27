## /send
A request to send a token to an address.
```json
{
  "service": "abracadabra",      // Service name (to differentiate multiple requestors): 1..64
  "id": "000001",                // Unique request ID (within service): 1..64
  "public_key": "YeAHC***pa97",  // Destination wallet address in Base58
  "token": "mnt",                // GOLD or MNT
  "amount": "1.666"              // Token amount in major units: 1.234 (18 decimal places)
}
```

---

## Notification
An event with sending completion status.
```json
{
  "success": true,               // Success is true in case of success
  "error": "",                   // Error contains error descrition in case of failure
  "service": "abracadabra",      // Service name (to differentiate multiple requestors): 1..64
  "id": "000001",                // Unique request ID: 1..64
  "public_key": "YeAHC***pa97",  // Destination wallet address in Base58 (empty on failure)
  "token": "mnt",                // GOLD or MNT (empty on failure)
  "amount": "1.666",             // Token amount in major units: 1.234 (18 decimal places, empty on failure)
  "transaction": "YeAHC***pa97"  // Transaction digest in Base58 (empty on failure)
}
```