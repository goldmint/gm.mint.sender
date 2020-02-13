package blockranger

import (
	"math/big"
	"time"

	"github.com/void616/gotask"
)

// Task loop
func (r *Ranger) Task(token *gotask.Token) {
	r.logger.Infof("Processing range from %v to %v", r.from.String(), r.to.String())

	cur := new(big.Int).Set(r.from)
	one := big.NewInt(1)

	for !token.Stopped() {

		if err := r.parser.Parse(cur); err != nil {
			r.logger.WithError(err).WithField("block", cur.String()).Error("Failed to parse block")
			token.Sleep(time.Second * 10)
			continue
		}
		r.logger.WithField("block", cur.String()).Debugf("Block completed")

		cur.Add(cur, one)
		if cur.Cmp(r.to) > 0 {
			r.logger.Infof("Range from %v to %v is completed", r.from.String(), r.to.String())
			token.Stop()
			continue
		}
	}
}
