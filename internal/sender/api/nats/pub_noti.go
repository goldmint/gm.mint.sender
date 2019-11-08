package nats

import (
	"fmt"
	"time"

	proto "github.com/golang/protobuf/proto"
	senderNatsProto "github.com/void616/gm-mint-sender/pkg/sender/nats"
	sumuslib "github.com/void616/gm-sumuslib"
	"github.com/void616/gm-sumuslib/amount"
)

// PublishSentEvent sends a sending completion notification
func (n *Nats) PublishSentEvent(
	success bool,
	msgerr string,
	service, requestID string,
	to sumuslib.PublicKey,
	token sumuslib.Token,
	amo *amount.Amount,
	digest *sumuslib.Digest,
) error {
	// metrics
	if n.metrics != nil {
		defer func(t time.Time) {
			n.metrics.NotificationDuration.Observe(time.Since(t).Seconds())
		}(time.Now())
	}

	transaction := ""
	if digest != nil {
		transaction = (*digest).String()
	}

	reqModel := senderNatsProto.Sent{
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

	msg, err := n.natsConnection.Request(n.subjPrefix+senderNatsProto.Sent{}.Subject(), req, time.Second*10)
	if err != nil {
		return err
	}

	repModel := senderNatsProto.SentAck{}
	if err := proto.Unmarshal(msg.Data, &repModel); err != nil {
		return err
	}

	if !repModel.GetSuccess() {
		return fmt.Errorf("service rejecttion: %v", repModel.GetError())
	}

	return nil
}
