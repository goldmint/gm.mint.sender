package blockobserver

import (
	"math/big"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"github.com/void616/gm-mint-sender/internal/blockparser"
	"github.com/void616/gm-mint-sender/internal/rpcpool"
)

// Observer listens for fresh blocks on an RPC connection
type Observer struct {
	logger  *logrus.Entry
	rpcpool *rpcpool.Pool
	parser  *blockparser.Parser
	from    *big.Int
	pubTX   chan<- *blockparser.Transaction
	blockID chan<- *big.Int

	mtxTaskDuration *prometheus.SummaryVec
	mtxQueueGauge   *prometheus.GaugeVec
}

// New Observer instance
func New(
	from *big.Int,
	pool *rpcpool.Pool,
	pubTX chan<- *blockparser.Transaction,
	blockID chan<- *big.Int,
	mtxTaskDuration *prometheus.SummaryVec,
	mtxQueueGauge *prometheus.GaugeVec,
	logger *logrus.Entry,
) (*Observer, error) {

	parser, err := blockparser.New(pool, pubTX, blockID, mtxTaskDuration)
	if err != nil {
		return nil, err
	}

	o := &Observer{
		rpcpool:         pool,
		logger:          logger,
		parser:          parser,
		from:            from,
		mtxTaskDuration: mtxTaskDuration,
		mtxQueueGauge:   mtxQueueGauge,
	}
	return o, nil
}
