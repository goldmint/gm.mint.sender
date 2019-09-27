package nats

import (
	"time"
	"unicode/utf8"

	"github.com/void616/gm-mint-sender/internal/sender/api/model"

	proto "github.com/golang/protobuf/proto"
	gonats "github.com/nats-io/go-nats"
	senderNats "github.com/void616/gm-mint-sender/pkg/sender/nats/sender"
	sumuslib "github.com/void616/gm-sumuslib"
	"github.com/void616/gm-sumuslib/amount"
)

// subSendRequest listens for a new sending requests until connection draining
func (n *Nats) subSendRequest(m *gonats.Msg) {
	nc := n.natsConnection

	// metrics
	if n.metrics != nil {
		defer func(t time.Time, method string) {
			n.metrics.RequestDuration.WithLabelValues("send").Observe(time.Since(t).Seconds())
		}(time.Now(), m.Subject)
	}

	// parse
	req := senderNats.SendRequest{}
	if err := proto.Unmarshal(m.Data, &req); err != nil {
		n.logger.WithError(err).Error("Failed to unmarshal request")
		return
	}

	n.logger.WithField("data", req.String()).Debug("Got sending request")

	// reply
	var replyError string
	defer func() {
		rep := senderNats.SendReply{
			Success: replyError == "",
			Error:   replyError,
		}
		if b, err := proto.Marshal(&rep); err != nil {
			n.logger.WithError(err).Error("Failed to marshal reply")
		} else {
			if err := nc.Publish(m.Reply, b); err != nil {
				n.logger.WithError(err).Error("Failed to publish reply")
			}
		}
	}()

	// check req service
	if x := utf8.RuneCountInString(req.GetService()); x < 1 || x > 64 || !model.ServiceNameRex.MatchString(req.GetService()) {
		replyError = "Invalid service name"
		return
	}

	// check req id
	if x := utf8.RuneCountInString(req.GetId()); x < 1 || x > 64 {
		replyError = "Invalid request ID"
		return
	}

	// parse wallet address
	reqAddr, err := sumuslib.ParsePublicKey(req.GetPublicKey())
	if err != nil {
		replyError = "Invalid public key"
		return
	}

	// parse token
	reqToken, err := sumuslib.ParseToken(req.GetToken())
	if err != nil {
		replyError = "Invalid token"
		return
	}

	// parse amount
	reqAmount, err := amount.FromString(req.GetAmount())
	if err != nil {
		replyError = "Invalid amount"
		return
	}

	// enqueue
	if dups, ok := n.api.EnqueueSendingNats(req.GetId(), req.GetService(), reqAddr, reqAmount, reqToken); !ok {
		if dups {
			replyError = "Request with the same ID registered"
		} else {
			replyError = "Internal failure"
		}
		return
	}
}
