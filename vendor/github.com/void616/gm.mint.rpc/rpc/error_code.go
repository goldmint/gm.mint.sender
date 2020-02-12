package rpc

// ErrorCode is a Mint node error code
type ErrorCode uint16

// General errors
const (
	// ESuccess - success
	ESuccess ErrorCode = 0
	// EUnclassified - unclassified error
	EUnclassified ErrorCode = 1
)

// JSON request errors
const (
	// EInvalidJSON - invalid json
	EInvalidJSON ErrorCode = 10
	// EMalformedJSONRPC - method, id or params field not specified
	EMalformedJSONRPC ErrorCode = 11
	// EMethodNotFound - method not found
	EMethodNotFound ErrorCode = 12
	// EMalformedRequest - mandatory field not specified
	EMalformedRequest ErrorCode = 13
)

// Blockchain manager errors
const (
	// EGetBlockFailure - get block failure
	EGetBlockFailure ErrorCode = 20
	// EBlockNotFound - block not found in DB
	EBlockNotFound ErrorCode = 21
)

// Synchronization errors
const (
	// EConsensusIsAlreadyStarted - consensus is already started
	EConsensusIsAlreadyStarted ErrorCode = 31
	// ESynchronizationIsAlreadyStarted - synchronization is already started
	ESynchronizationIsAlreadyStarted ErrorCode = 32
	// EBadNodeCount - bad node count: number of nodes is greater than number of nodes for voting
	EBadNodeCount ErrorCode = 33
	// EUnknownNode - specified node for manual synchronization is not found in blockchain
	EUnknownNode ErrorCode = 34
)

// Transaction pool errors
const (
	// ETransactionNotSigned - transaction is not signed
	ETransactionNotSigned ErrorCode = 40
	// EBadTransactionSignature - bad transaction signature
	EBadTransactionSignature ErrorCode = 41
	// EVotingPoolOverflow - voting pool overflow
	EVotingPoolOverflow ErrorCode = 42
	// EPendingPoolOverflow - pending pool overflow
	EPendingPoolOverflow ErrorCode = 43
	// ETransactionWalletNotFound - existing transaction wallet not found
	ETransactionWalletNotFound ErrorCode = 44
	// EBadTransactionID - transaction ID is leser or equal to last transaction ID in last approved block
	EBadTransactionID ErrorCode = 45
	// EBadTransactionDeltaID - delta ID is exceeded doubled max size of the block
	EBadTransactionDeltaID ErrorCode = 46
	// EBadTransaction - bad transaction that cannot be applied to its wallet
	EBadTransaction ErrorCode = 47
	// ETransactionIDExistsInVotingPool - transaction with specified ID already exists in voting pool
	ETransactionIDExistsInVotingPool ErrorCode = 48
	// ETransactionIDExistsInPendingPool - transaction with specified ID already exists in pending pool
	ETransactionIDExistsInPendingPool ErrorCode = 49
	// EMaxSizeOfTransactionPackExceeded - max size of transaction pack exceeded
	EMaxSizeOfTransactionPackExceeded ErrorCode = 50
)

// String is stringer interface
func (ec ErrorCode) String() string {
	if str, ok := errorCodeDesc[ec]; ok {
		return str
	}
	return "error code not implemented"
}

var errorCodeDesc = map[ErrorCode]string{
	// general
	ESuccess:      "success",
	EUnclassified: "unclassified",
	// json
	EInvalidJSON:      "jsonrpc: invalid json",
	EMalformedJSONRPC: "jsonrpc: method, id or params not specified",
	EMethodNotFound:   "jsonrpc: method not found",
	EMalformedRequest: "jsonrpc: mandatory field not specified",
	// bc manager
	EGetBlockFailure: "blockchain: get block failure",
	EBlockNotFound:   "blockchain: block not found",
	// sync
	EConsensusIsAlreadyStarted:       "sync: consensus is already started",
	ESynchronizationIsAlreadyStarted: "sync: sync is already started",
	EBadNodeCount:                    "sync: number of nodes is greater than number of nodes for voting",
	EUnknownNode:                     "sync: given node for manual sync not found",
	// tx pool
	ETransactionNotSigned:             "txpool: tx is not signed",
	EBadTransactionSignature:          "txpool: bad tx signature",
	EVotingPoolOverflow:               "txpool: voting pool overflow",
	EPendingPoolOverflow:              "txpool: pending pool overflow",
	ETransactionWalletNotFound:        "txpool: existing wallet not found",
	EBadTransactionID:                 "txpool: tx id is lte to latest approved tx id",
	EBadTransactionDeltaID:            "txpool: tx delta id is exceeded doubled max size of the block",
	EBadTransaction:                   "txpool: tx can't be applied to the wallet",
	ETransactionIDExistsInVotingPool:  "txpool: tx id already exists in voting pool",
	ETransactionIDExistsInPendingPool: "txpool: tx id already exists in pending pool",
	EMaxSizeOfTransactionPackExceeded: "txpool: max size of tx pack exceeded",
}

// ---

// TxAddedAlready returns true in case the transaction is already in pending/voting pool (added before)
func (ec ErrorCode) TxAddedAlready() bool {
	return ec == ETransactionIDExistsInVotingPool || ec == ETransactionIDExistsInPendingPool
}

// TxNonceBehind returns true in case sent transaction has wrong ID (ID <= latest approved ID)
func (ec ErrorCode) TxNonceBehind() bool {
	return ec == EBadTransactionID
}

// TxNonceAhead returns true in case sent transaction has wrong ID (ID >> ID allowed for this wallet in this block)
func (ec ErrorCode) TxNonceAhead() bool {
	return ec == EBadTransactionDeltaID
}

// TxWalletNotReady returns true in case the sending wallet doesn't exist or hasn't enough funds/tags for the transaction
func (ec ErrorCode) TxWalletNotReady() bool {
	return ec == EBadTransaction || ec == ETransactionWalletNotFound
}
