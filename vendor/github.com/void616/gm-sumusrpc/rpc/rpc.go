package rpc

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"unicode"

	sumuslib "github.com/void616/gm-sumuslib"
	"github.com/void616/gm-sumuslib/amount"
	"github.com/void616/gm-sumusrpc/conn"
)

// AddTransaction posts a new transaction
func AddTransaction(c *conn.Conn, t sumuslib.Transaction, hexdata string) (result AddTransactionResult, code ErrorCode, err error) {
	code, err = ECUnclassified, nil
	result = AddTransactionResult{}

	req := struct {
		TransactionName string `json:"transaction_name,omitempty"`
		TransactionData string `json:"transaction_data,omitempty"`
	}{
		t.String(),
		hexdata,
	}
	res := struct {
		AddedToVotingPool   int    `json:"added_to_voting_pool,string,omitempty"`
		VotingPoolCapacity  uint32 `json:"number_of_remaining_voting_transactions,string,omitempty"`
		PendingPoolCapacity uint32 `json:"number_of_remaining_pending_transactions,string,omitempty"`
	}{}

	code, err = RawCall(c, "add-transaction", &req, &res)
	if code != ECSuccess || err != nil {
		return
	}

	result.AddedToVotingPool = res.AddedToVotingPool == 1
	result.VotingPoolCapacity = res.VotingPoolCapacity
	result.PendingPoolCapacity = res.PendingPoolCapacity
	return
}

// WalletState gets state of specific wallet
func WalletState(c *conn.Conn, address string) (state WalletStateResult, code ErrorCode, err error) {
	code, err = ECUnclassified, nil
	state = WalletStateResult{
		Balance: WalletBalanceResult{
			Gold: amount.FromInteger(0),
			Mnt:  amount.FromInteger(0),
		},
	}

	req := struct {
		PublicKey string `json:"public_key,omitempty"`
	}{
		address,
	}
	res := struct {
		Balance           json.RawMessage `json:"balance,omitempty"`
		Exist             int             `json:"exist,string,omitempty"`
		LastTransactionID uint64          `json:"last_transaction_id,string,omitempty"`
		Tags              json.RawMessage `json:"tags,omitempty"`
	}{}
	type BalanceItem struct {
		AssetCode string `json:"asset_code,omitempty"`
		Amount    string `json:"amount,omitempty"`
	}

	code, err = RawCall(c, "get-wallet-state", &req, &res)
	if code != ECSuccess || err != nil {
		return
	}

	// balance exists
	if res.Balance == nil {
		err = fmt.Errorf("balance field not set")
		return
	}
	// balance is array or string - try parse as array
	balance := []BalanceItem{}
	if perr := json.Unmarshal(res.Balance, &balance); perr == nil {
		for _, v := range balance {
			if parsed, err := amount.FromString(v.Amount); err == nil {
				if token, perr := sumuslib.ParseToken(v.AssetCode); perr == nil {
					switch token {
					case sumuslib.TokenMNT:
						state.Balance.Mnt = parsed
					case sumuslib.TokenGOLD:
						state.Balance.Gold = parsed
					}
				}
			}
		}
	}
	// tags is array or string - try parse as array
	tags := []string{}
	json.Unmarshal(res.Tags, &tags)
	state.Exists = res.Exist == 1
	state.ApprovedNonce = res.LastTransactionID
	state.Tags = tags
	return
}

// BlockchainState gets blockchain state info
func BlockchainState(c *conn.Conn) (state BlockchainStateResult, code ErrorCode, err error) {
	code, err = ECUnclassified, nil
	state = BlockchainStateResult{
		BlockCount:              new(big.Int),
		LastBlockDigest:         "",
		LastBlockMerkleRoot:     "",
		TransactionCount:        new(big.Int),
		NodeCount:               new(big.Int),
		NonEmptyWalletCount:     new(big.Int),
		VotingTransactionCount:  new(big.Int),
		PendingTransactionCount: new(big.Int),
		BlockchainState:         "",
		ConsensusRound:          "",
		VotingNodes:             "",
		Balance:                 BlockchainBalanceResult{},
	}

	req := struct{}{}
	res := struct {
		BlockCount              string `json:"block_count,omitempty"`
		LastBlockDigest         string `json:"last_block_digest,omitempty"`
		LastBlockMerkleRoot     string `json:"last_block_merkle_root,omitempty"`
		TransactionCount        string `json:"transaction_count,omitempty"`
		NodeCount               string `json:"node_count,omitempty"`
		NonEmptyWalletCount     string `json:"non_empty_wallet_count,omitempty"`
		VotingTransactionCount  string `json:"voing_transaction_count,omitempty"`
		PendingTransactionCount string `json:"pending_transaction_count,omitempty"`
		BlockchainState         string `json:"blockchain_state,omitempty"`
		ConsensusRound          string `json:"consensus_round,omitempty"`
		VotingNodes             string `json:"voting_nodes,omitempty"`
		Balance                 struct {
			Mnt  *amount.Amount `json:"mnt,omitempty"`
			Gold *amount.Amount `json:"gold,omitempty"`
		} `json:"balance,omitempty"`
	}{}

	code, err = RawCall(c, "get-blockchain-state", &req, &res)
	if code != ECSuccess || err != nil {
		return
	}

	// blocks count
	if i, ok := new(big.Int).SetString(res.BlockCount, 10); ok {
		state.BlockCount = i
	} else {
		err = errors.New("failed to parse blocks count")
		return
	}
	// tx count
	if i, ok := new(big.Int).SetString(res.TransactionCount, 10); ok {
		state.TransactionCount = i
	} else {
		err = errors.New("failed to parse transactions count")
		return
	}
	// nodes count
	if i, ok := new(big.Int).SetString(res.NodeCount, 10); ok {
		state.NodeCount = i
	} else {
		err = errors.New("failed to parse nodes count")
		return
	}
	// non-empty wallets
	if i, ok := new(big.Int).SetString(res.NonEmptyWalletCount, 10); ok {
		state.NonEmptyWalletCount = i
	} else {
		err = errors.New("failed to parse non-empty wallets count")
		return
	}
	// voting transactions
	if i, ok := new(big.Int).SetString(res.VotingTransactionCount, 10); ok {
		state.VotingTransactionCount = i
	} else {
		err = errors.New("failed to parse voting transactions count")
		return
	}
	// pending transactions
	if i, ok := new(big.Int).SetString(res.PendingTransactionCount, 10); ok {
		state.PendingTransactionCount = i
	} else {
		err = errors.New("failed to parse pending transactions count")
		return
	}
	// other
	state.LastBlockDigest = res.LastBlockDigest
	state.LastBlockMerkleRoot = res.LastBlockMerkleRoot
	state.BlockchainState = res.BlockchainState
	state.ConsensusRound = res.ConsensusRound
	state.VotingNodes = res.VotingNodes
	state.Balance.Mnt = res.Balance.Mnt
	state.Balance.Gold = res.Balance.Gold
	return
}

// BlockData gets raw block data (hex) by ID
func BlockData(c *conn.Conn, id *big.Int) (data string, code ErrorCode, err error) {
	code, err = ECUnclassified, nil
	data = ""

	req := struct {
		BlockID string `json:"block_id,omitempty"`
	}{
		id.String(),
	}
	res := struct {
		BlockData string `json:"block_data,omitempty"`
	}{}

	code, err = RawCall(c, "get-block", &req, &res)
	if code != ECSuccess || err != nil {
		return
	}

	data = res.BlockData
	return
}

// Nodes gets blockchain nodes list
func Nodes(c *conn.Conn) (nodes []NodeResult, code ErrorCode, err error) {
	code, err = ECUnclassified, nil
	nodes = make([]NodeResult, 0)

	req := struct{}{}
	type Item struct {
		Index   string `json:"index,omitempty"`
		Wallet  string `json:"wallet_public_key,omitempty"`
		Address string `json:"address,omitempty"`
	}
	res := struct {
		NodeList []Item `json:"node_list,omitempty"`
	}{}

	code, err = RawCall(c, "get-nodes", &req, &res)
	if code != ECSuccess || err != nil {
		return
	}

	// list exists
	if res.NodeList == nil || len(res.NodeList) == 0 {
		err = fmt.Errorf("node list is empty")
		return
	}
	for _, v := range res.NodeList {
		nodes = append(nodes, NodeResult{
			Index:   v.Index,
			Address: v.Wallet,
			IP:      v.Address,
		})
	}
	return
}

// WalletTransactions returns wallet incoming/outgoing transaction list
func WalletTransactions(c *conn.Conn, count uint16, address string) (list []WalletTransactionsResult, code ErrorCode, err error) {
	code, err = ECUnclassified, nil
	list = make([]WalletTransactionsResult, 0)

	req := struct {
		PublicKey string `json:"public_key,omitempty"`
		Count     string `json:"count,omitempty"`
	}{
		address, fmt.Sprintf("%v", count),
	}
	type Item struct {
		Description string `json:"description,omitempty"`
		Digest      string `json:"digest,omitempty"`
		Type        string `json:"type,omitempty"`
	}
	res := struct {
		TxList []Item `json:"transaction_list,omitempty"`
	}{}

	code, err = RawCall(c, "get-wallet-transactions", &req, &res)
	if code != ECSuccess || err != nil {
		return
	}

	for _, v := range res.TxList {
		from, to, nonce, dummy := "", "", uint64(0), ""

		rarr := []rune(v.Description)
		for i, v := range rarr {
			if !unicode.IsDigit(v) && !unicode.IsLetter(v) && v != '.' {
				rarr[i] = ' '
			}
		}
		// TransferAssetsTransaction ID 65  Q8Uz1RmhXF6PCw2RAGKUCR8yQMAaMJDGSxPkKoyfAvr3svs41  247jCgpxw5VMZsXe1kPHuFpeVdWGqUDpfmDznG2ebNar8hzWYq  19.096357648835176946 mnt
		n, serr := fmt.Sscanf(
			string(rarr),
			"%s %s %d %s %s",
			&dummy, &dummy, &nonce, &from, &to,
		)
		if serr != nil {
			err = serr
			return
		}
		if n != 5 {
			err = errors.New("failed to parse description")
			return
		}

		list = append(list, WalletTransactionsResult{
			From:   from,
			To:     to,
			Nonce:  nonce,
			Digest: v.Digest,
			Status: v.Type,
		})
	}
	return
}
