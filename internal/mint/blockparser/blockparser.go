package blockparser

import (
	"math/big"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/void616/gm.mint.sender/internal/mint/rpcpool"
)

// Parser parses specific block data on demand and extracts transactions
type Parser struct {
	rpcpool *rpcpool.Pool
	pubTX   chan<- *Transaction
	blockID chan<- *big.Int
	metrics *Metrics
}

// New Parser instance
func New(
	rpcpool *rpcpool.Pool,
	pubTX chan<- *Transaction,
	blockID chan<- *big.Int,
) (*Parser, error) {
	ret := &Parser{
		rpcpool: rpcpool,
		pubTX:   pubTX,
		blockID: blockID,
	}
	return ret, nil
}

// Metrics data
type Metrics struct {
	RequestDuration prometheus.Histogram
	ParsingDuration prometheus.Histogram
}

// AddMetrics adds metrics counters and should be called before service launch
func (p *Parser) AddMetrics(m *Metrics) {
	p.metrics = m
}
