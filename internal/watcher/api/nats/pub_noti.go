package nats

import (
	"fmt"
	"time"

	proto "github.com/golang/protobuf/proto"
	walletsvc "github.com/void616/gm.mint.sender/pkg/watcher/nats"
	mint "github.com/void616/gm.mint"
	"github.com/void616/gm.mint/amount"
)

// NotifyRefilling sends an event
func (n *Nats) NotifyRefilling(service string, to, from mint.PublicKey, t mint.Token, a *amount.Amount, tx mint.Digest) error {

	// metrics
	if n.metrics != nil {
		defer func(t time.Time) {
			n.metrics.NotificationDuration.Observe(time.Since(t).Seconds())
		}(time.Now())
	}

	reqModel := &walletsvc.Refill{
		Service:     service,
		PublicKey:   to.String(),
		From:        from.String(),
		Token:       t.String(),
		Amount:      a.String(),
		Transaction: tx.String(),
	}

	req, err := proto.Marshal(reqModel)
	if err != nil {
		return err
	}

	msg, err := n.natsConnection.Request(n.subjPrefix+walletsvc.Refill{}.Subject(), req, time.Second*5)
	if err != nil {
		return err
	}

	repModel := walletsvc.RefillAck{}
	if err := proto.Unmarshal(msg.Data, &repModel); err != nil {
		return err
	}

	if !repModel.GetSuccess() {
		return fmt.Errorf("service rejection: %v", repModel.GetError())
	}

	return nil
}
