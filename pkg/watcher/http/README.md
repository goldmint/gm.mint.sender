## /watch
A request to add wallet to observation.
```json
{
  "service": "abracadabra",                   // Service name (to differentiate multiple requestors): 1..64
  "public_keys": ["YeAHC...pa97"],            // Destination wallet address in Base58
  "callback": "https://example.com/callback"  // Callback for notification: 1..256
}
```

---
## /unwatch
A request to remove wallet from observation.
```json
{
  "service": "abracadabra",        // Service name (to differentiate multiple requestors): 1..64
  "public_keys": ["YeAHC...pa97"]  // Destination wallet address in Base58
}
```

---

## POST callback (notification)
An event from watcher.
```json
{
  "service": "abracadabra",      // Service name (to differentiate multiple requestors): 1..64
  "public_key": "YeAHC...pa97",  // Destination (watching) wallet address in Base58
  "from": "Fr0M...ad12",         // Source wallet address in Base58
  "token": "GOLD",               // GOLD or MNT
  "amount": "1.666...000",       // Token amount in major units (18 decimal places)
  "transaction": "DgS1...1234",  // Digest of the refilling tx in Base58
}
```
