package txsigner

import (
	"fmt"
	"math/big"
	"time"

	"github.com/void616/gm-mint-sender/internal/sender/db/types"
	"github.com/void616/gm-sumusrpc/rpc"
	"github.com/void616/gotask"
)

// Task loop
func (s *Signer) Task(token *gotask.Token) {

	requests := make(chan *types.Sending, itemsPerShot*2)
	defer close(requests)

	currentBlock := new(big.Int)

	for !token.Stopped() {

		// get current network block
		{
			conn, err := s.pool.Get()
			if err != nil {
				s.logger.WithError(err).Error("Failed to get RPC connection")
				token.Sleep(time.Second * 30)
				continue
			}
			state, code, err := rpc.BlockchainState(conn.Conn())
			if err != nil || code != rpc.ECSuccess {
				conn.Close()
				if code != rpc.ECSuccess {
					err = fmt.Errorf("node code %v", code)
				}
				s.logger.WithError(err).Error("Failed to get current block ID")
				token.Sleep(time.Second * 30)
				continue
			}
			currentBlock.Sub(state.BlockCount, big.NewInt(1))
			conn.Close()
		}

		count := 0

		// get stale requests
		{
			elderThan := new(big.Int).Sub(currentBlock, big.NewInt(staleAfterBlocks))

			list, err := s.dao.ListStaleSendings(elderThan, itemsPerShot)
			if err != nil {
				s.logger.WithError(err).Error("Failed to get stale transactions")
				token.Sleep(time.Second * 30)
				continue
			}
			for _, v := range list {
				requests <- v
			}
			count += len(list)
		}

		// get new requests
		{
			list, err := s.dao.ListEnqueuedSendings(itemsPerShot)
			if err != nil {
				s.logger.WithError(err).Error("Failed to get new transactions")
				token.Sleep(time.Second * 30)
				continue
			}
			for _, v := range list {
				requests <- v
			}
			count += len(list)
		}

		// empty queue
		if count == 0 {
			token.Sleep(time.Second * 10)
			continue
		}

		// metrics
		if s.mtxQueueGauge != nil {
			s.mtxQueueGauge.WithLabelValues("txsigner_shot").Set(float64(count))
		}

		// process queue
		processed := 0
		out := false
		for !out {
			select {
			default:
				out = true
			case snd := <-requests:

				// metrics
				t := time.Now()

				if s.processRequest(snd, currentBlock) {
					processed++
				}

				// metrics
				if s.mtxTaskDuration != nil {
					s.mtxTaskDuration.WithLabelValues("txsigner_shot").Observe(time.Since(t).Seconds())
				}
			}
		}

		// metrics
		if s.mtxQueueGauge != nil {
			s.mtxQueueGauge.WithLabelValues("txsigner_shot").Set(0)
		}

		if processed == 0 {
			token.Sleep(time.Second * 30)
		}
	}
}
