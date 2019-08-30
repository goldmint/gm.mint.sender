package nats

import (
	"fmt"
	"time"

	proto "github.com/golang/protobuf/proto"
	walletsvc "github.com/void616/gm-mint-sender/pkg/watcher/nats/wallet"
	sumuslib "github.com/void616/gm-sumuslib"
	"github.com/void616/gm-sumuslib/amount"
)

// NotifyRefilling sends event
func (s *Service) NotifyRefilling(service string, addr sumuslib.PublicKey, t sumuslib.Token, a *amount.Amount, tx sumuslib.Digest) error {

	// metrics
	mt := time.Now()

	reqModel := &walletsvc.RefillEvent{
		Service:     service,
		PublicKey:   addr.String(),
		Token:       t.String(),
		Amount:      a.String(),
		Transaction: tx.String(),
	}

	req, err := proto.Marshal(reqModel)
	if err != nil {
		return err
	}

	msg, err := s.natsConnection.Request(s.subjPrefix+walletsvc.SubjectRefill, req, time.Second*5)
	if err != nil {
		return err
	}

	repModel := walletsvc.RefillEventReply{}
	if err := proto.Unmarshal(msg.Data, &repModel); err != nil {
		return err
	}

	if !repModel.GetSuccess() {
		return fmt.Errorf("service rejection: %v", repModel.GetError())
	}

	// metrics
	if s.mtxTaskDuration != nil {
		s.mtxTaskDuration.WithLabelValues("nats_notifier").Observe(time.Since(mt).Seconds())
	}

	return nil
}
