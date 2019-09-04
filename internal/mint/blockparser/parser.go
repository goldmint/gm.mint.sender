package blockparser

import (
	"math/big"
	"time"
)

// Parse requires and parses specified block by ID
func (p *Parser) Parse(block *big.Int) error {
	// metrics
	t := time.Now()

	// require block
	blockBytes, err := p.queryBlockData(block)
	if err != nil {
		return err
	}
	defer blockBytes.Close()

	// metrics
	if p.metrics != nil {
		p.metrics.RequestDuration.Observe(time.Since(t).Seconds())
	}

	t = time.Now()
	if err := p.parseBlockData(blockBytes); err != nil {
		return err
	}

	// metrics
	if p.metrics != nil {
		p.metrics.ParsingDuration.Observe(time.Since(t).Seconds())
	}

	// send as parsed
	p.blockID <- new(big.Int).Set(block)
	return nil
}
