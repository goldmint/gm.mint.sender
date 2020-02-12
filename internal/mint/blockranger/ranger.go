package blockranger

import (
	"errors"
	"math/big"

	"github.com/sirupsen/logrus"
	"github.com/void616/gm.mint.sender/internal/mint/blockparser"
	"github.com/void616/gm.mint.sender/internal/mint/rpcpool"
)

// Ranger parses range of blocks sending IDs to the parsers channel
type Ranger struct {
	logger *logrus.Entry
	from   *big.Int
	to     *big.Int
	parser *blockparser.Parser
	pubTX  chan<- *blockparser.Transaction
}

// New Ranger instance.
// Parses blocks from `fromBlockID` to `toBlockID` (inclusive).
func New(
	fromBlockID,
	toBlockID *big.Int,
	pool *rpcpool.Pool,
	pubTX chan<- *blockparser.Transaction,
	blockID chan<- *big.Int,
	logger *logrus.Entry,
) (*Ranger, error) {

	from, to := big.NewInt(0), big.NewInt(0)
	if fromBlockID == nil || fromBlockID.Cmp(new(big.Int)) < 0 {
		return nil, errors.New("invalid range")
	}
	from.Set(fromBlockID)
	if toBlockID == nil || toBlockID.Cmp(new(big.Int)) < 0 {
		return nil, errors.New("invalid range")
	}
	to.Set(toBlockID)
	if from.Cmp(to) > 0 {
		return nil, errors.New("invalid range")
	}

	parser, err := blockparser.New(pool, pubTX, blockID)
	if err != nil {
		return nil, err
	}

	r := &Ranger{
		logger: logger,
		from:   from,
		to:     to,
		parser: parser,
	}
	return r, nil
}

// AddMetrics adds metrics counters and should be called before service launch
func (r *Ranger) AddMetrics(parser *blockparser.Metrics) {
	r.parser.AddMetrics(parser)
}
