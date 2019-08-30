package nats

import (
	"fmt"
	"time"

	proto "github.com/golang/protobuf/proto"
	senderNatsProto "github.com/void616/gm-mint-sender/pkg/sender/nats/sender"
	sumuslib "github.com/void616/gm-sumuslib"
	"github.com/void616/gm-sumuslib/amount"
)

// PublishSentEvent sends a "sent" event
func (s *Service) PublishSentEvent(
	success bool,
	msgerr string,
	service, requestID string,
	to sumuslib.PublicKey,
	token sumuslib.Token,
	amo *amount.Amount,
	digest *sumuslib.Digest,
) error {
	// metrics
	mt := time.Now()

	transaction := ""
	if digest != nil {
		transaction = (*digest).String()
	}

	reqModel := senderNatsProto.SentEvent{
		Success:     success,
		Error:       msgerr,
		Service:     service,
		Id:          requestID,
		PublicKey:   to.String(),
		Token:       token.String(),
		Amount:      amo.String(),
		Transaction: transaction,
	}

	req, err := proto.Marshal(&reqModel)
	if err != nil {
		return err
	}

	msg, err := s.natsConnection.Request(s.subjPrefix+senderNatsProto.SubjectSent, req, time.Second*5)
	if err != nil {
		return err
	}

	repModel := senderNatsProto.SentEventReply{}
	if err := proto.Unmarshal(msg.Data, &repModel); err != nil {
		return err
	}

	if !repModel.GetSuccess() {
		return fmt.Errorf("service rejecttion: %v", repModel.GetError())
	}

	// metrics
	if s.mtxTaskDuration != nil {
		s.mtxTaskDuration.WithLabelValues("nats_notifier").Observe(time.Since(mt).Seconds())
	}

	return nil
}
