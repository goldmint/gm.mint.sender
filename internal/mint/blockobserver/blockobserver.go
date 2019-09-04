package blockobserver

import (
	"math/big"

	"github.com/sirupsen/logrus"
	"github.com/void616/gm-mint-sender/internal/mint/blockparser"
	"github.com/void616/gm-mint-sender/internal/mint/rpcpool"
)

// Observer listens for fresh blocks on an RPC connection
type Observer struct {
	logger  *logrus.Entry
	rpcpool *rpcpool.Pool
	parser  *blockparser.Parser
	from    *big.Int
	pubTX   chan<- *blockparser.Transaction
	blockID chan<- *big.Int
}

// New Observer instance
func New(
	from *big.Int,
	pool *rpcpool.Pool,
	pubTX chan<- *blockparser.Transaction,
	blockID chan<- *big.Int,
	logger *logrus.Entry,
) (*Observer, error) {

	parser, err := blockparser.New(pool, pubTX, blockID)
	if err != nil {
		return nil, err
	}

	o := &Observer{
		rpcpool: pool,
		logger:  logger,
		parser:  parser,
		from:    from,
	}
	return o, nil
}

// AddMetrics adds metrics counters and should be called before service launch
func (o *Observer) AddMetrics(parser *blockparser.Metrics) {
	o.parser.AddMetrics(parser)
}
