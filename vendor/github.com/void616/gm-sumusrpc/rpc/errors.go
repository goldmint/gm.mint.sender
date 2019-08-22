package rpc

// ErrorCode is an RPC error from a node
type ErrorCode uint16

// String is stringer interface
func (ec ErrorCode) String() string {
	if str, ok := errorCodeDesc[ec]; ok {
		return str
	}
	return errorCodeDesc[ECUnclassified]
}

// RPC client errors
const (

	// General errors:

	// ECSuccess: success
	ECSuccess ErrorCode = 0
	// ECUnclassified: unclassified error
	ECUnclassified ErrorCode = 1

	// JSON request errors:

	// ECJSONBadRequest: bad request that cannot be parsed
	ECJSONBadRequest ErrorCode = 10
	// ECJSONRequestIDNotFound: request with not specified field 'id'
	ECJSONRequestIDNotFound ErrorCode = 11
	// ECJSONUnknownDebugRequest: debug request with unknown ID
	ECJSONUnknownDebugRequest ErrorCode = 12
	// ECJSONUnknownJsonRequest: JSON request with unknown ID
	ECJSONUnknownJSONRequest ErrorCode = 13
	// ECJSONBadRequestFormat: bad JSON request format e.g. at least one mandatory field not found
	ECJSONBadRequestFormat ErrorCode = 14

	// Blockchain manager errors:

	// ECGetBlockFailure: get block failure
	ECGetBlockFailure ErrorCode = 20
	// ECBlockNotFound: block not found in DB
	ECBlockNotFound ErrorCode = 21

	// Synchronization errors:

	// ECConsensusIsAlreadyStarted: consensus is already started
	ECConsensusIsAlreadyStarted ErrorCode = 31
	// ECSynchronizationIsAlreadyStarted: synchronization is already started
	ECSynchronizationIsAlreadyStarted ErrorCode = 32
	// ECBadNodeCount: bad node count: number of nodes is greater than number of nodes for voting
	ECBadNodeCount ErrorCode = 33
	// ECUnknownNode: specified node for manual synchronization is not found in blockchain
	ECUnknownNode ErrorCode = 34

	// Transaction pool errors:

	// ECTransactionNotSigned: transaction is not signed
	ECTransactionNotSigned ErrorCode = 40
	// ECBadTransactionSignature: bad transaction signature
	ECBadTransactionSignature ErrorCode = 41
	// ECVotingPoolOverflow: voting pool overflow
	ECVotingPoolOverflow ErrorCode = 42
	// ECPendingPoolOverflow: pending pool overflow
	ECPendingPoolOverflow ErrorCode = 43
	// ECTransactionWalletNotFound: existing transaction wallet not found
	ECTransactionWalletNotFound ErrorCode = 44
	// ECBadTransactionID: transaction ID is leser or equal to last transaction ID in last approved block
	ECBadTransactionID ErrorCode = 45
	// ECBadTransactionDeltaID: delta ID is exceeded doubled max size of the block
	ECBadTransactionDeltaID ErrorCode = 46
	// ECBadTransaction: bad transaction that cannot be applied to its wallet
	ECBadTransaction ErrorCode = 47
	// ECTransactionIDExistsInVotingPool: transaction with specified ID already exists in voting pool
	ECTransactionIDExistsInVotingPool ErrorCode = 48
	// ECTransactionIDExistsInPendingPool: transaction with specified ID already exists in pending pool
	ECTransactionIDExistsInPendingPool ErrorCode = 49
	// ECMaxSizeOfTransactionPackExceeded: max size of transaction pack exceeded
	ECMaxSizeOfTransactionPackExceeded ErrorCode = 50
)

var errorCodeDesc = map[ErrorCode]string{
	ECSuccess:                          "Success",
	ECUnclassified:                     "Unclassified error",
	ECJSONBadRequest:                   "JSON request error: bad request that cannot be parsed",
	ECJSONRequestIDNotFound:            "JSON request error: request with not specified field 'id'",
	ECJSONUnknownDebugRequest:          "JSON request error: debug request with unknown ID",
	ECJSONUnknownJSONRequest:           "JSON request error: JSON request with unknown ID",
	ECJSONBadRequestFormat:             "JSON request error: bad JSON request format e.g. at least one mandatory field not found",
	ECGetBlockFailure:                  "Blockchain manager error: get block failure",
	ECBlockNotFound:                    "Blockchain manager error: block not found in DB",
	ECConsensusIsAlreadyStarted:        "Synchronization error: consensus is already started",
	ECSynchronizationIsAlreadyStarted:  "Synchronization error: synchronization is already started",
	ECBadNodeCount:                     "Synchronization error: bad node count, number of nodes is greater than number of nodes for voting",
	ECUnknownNode:                      "Synchronization error: specified node for manual synchronization is not found in blockchain",
	ECTransactionNotSigned:             "Transaction pool error: transaction is not signed",
	ECBadTransactionSignature:          "Transaction pool error: bad transaction signature",
	ECVotingPoolOverflow:               "Transaction pool error: voting pool overflow",
	ECPendingPoolOverflow:              "Transaction pool error: pending pool overflow",
	ECTransactionWalletNotFound:        "Transaction pool error: existing transaction wallet not found",
	ECBadTransactionID:                 "Transaction pool error: transaction ID is leser or equal to last transaction ID in last approved block",
	ECBadTransactionDeltaID:            "Transaction pool error: delta ID is exceeded doubled max size of the block",
	ECBadTransaction:                   "Transaction pool error: bad transaction that cannot be applied to its wallet",
	ECTransactionIDExistsInVotingPool:  "Transaction pool error: transaction with specified ID already exists in voting pool",
	ECTransactionIDExistsInPendingPool: "Transaction pool error: transaction with specified ID already exists in pending pool",
	ECMaxSizeOfTransactionPackExceeded: "Transaction pool error: max size of transaction pack exceeded",
}

// ---

// AddTransactionErrorCode is a special wrapper for RPC error
type AddTransactionErrorCode ErrorCode

// AddedAlready returns true in case the transaction is already in pending/voting pool (added before)
func (atec AddTransactionErrorCode) AddedAlready() bool {
	return ErrorCode(atec) == ECTransactionIDExistsInVotingPool || ErrorCode(atec) == ECTransactionIDExistsInPendingPool
}

// NonceBehind returns true in case sent transaction has wrong ID (ID <= latest approved ID)
func (atec AddTransactionErrorCode) NonceBehind() bool {
	return ErrorCode(atec) == ECBadTransactionID
}

// NonceAhead returns true in case sent transaction has wrong ID (ID >> ID allowed for this wallet in this block)
func (atec AddTransactionErrorCode) NonceAhead() bool {
	return ErrorCode(atec) == ECBadTransactionDeltaID
}

// WalletInconsistency returns true in case the sending wallet doesn't exist or hasn't enough funds/tags for the transaction
func (atec AddTransactionErrorCode) WalletInconsistency() bool {
	return ErrorCode(atec) == ECBadTransaction || ErrorCode(atec) == ECTransactionWalletNotFound
}
