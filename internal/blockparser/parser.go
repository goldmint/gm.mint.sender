package blockparser

import (
	"math/big"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/void616/gm-mint-sender/internal/rpcpool"
)

// Parser parses specific block data on demand and extracts transactions
type Parser struct {
	rpcpool *rpcpool.Pool
	pubTX   chan<- *Transaction

	mtxTaskDuration *prometheus.SummaryVec
}

// New Parser instance
func New(
	rpcpool *rpcpool.Pool,
	pubTX chan<- *Transaction,
	mtxTaskDuration *prometheus.SummaryVec,
) (*Parser, error) {
	ret := &Parser{
		rpcpool:         rpcpool,
		pubTX:           pubTX,
		mtxTaskDuration: mtxTaskDuration,
	}
	return ret, nil
}

// Parse requires and parses specified block by ID
func (p *Parser) Parse(block *big.Int) error {
	t := time.Now()
	blockBytes, err := p.queryBlockData(block)
	if err != nil {
		return err
	}
	defer blockBytes.Close()

	// metrics
	if p.mtxTaskDuration != nil {
		p.mtxTaskDuration.WithLabelValues("blockparser_query").Observe(time.Since(t).Seconds())
	}

	t = time.Now()
	if err := p.parseBlockData(blockBytes); err != nil {
		return err
	}

	// metrics
	if p.mtxTaskDuration != nil {
		p.mtxTaskDuration.WithLabelValues("blockparser_parse").Observe(time.Since(t).Seconds())
	}
	return nil
}
