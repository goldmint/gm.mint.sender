package nats

import (
	"time"

	proto "github.com/golang/protobuf/proto"
	gonats "github.com/nats-io/nats.go"
	mint "github.com/void616/gm.mint"
	"github.com/void616/gm.mint.sender/internal/sender/api/model"
	"github.com/void616/gm.mint.sender/internal/sender/db/types"
	senderNats "github.com/void616/gm.mint.sender/pkg/sender/nats"
	"github.com/void616/gm.mint/amount"
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
	req := senderNats.Send{}
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
	if !model.ServiceNameRex.MatchString(req.GetService()) {
		replyError = "invalid service name"
		return
	}

	// check req id
	if !model.RequestIDRex.MatchString(req.GetId()) {
		replyError = "invalid request ID"
		return
	}

	// parse wallet address
	reqAddr, err := mint.ParsePublicKey(req.GetPublicKey())
	if err != nil {
		replyError = "invalid public key"
		return
	}

	// parse token
	reqToken, err := mint.ParseToken(req.GetToken())
	if err != nil {
		replyError = "invalid token"
		return
	}

	// parse amount
	reqAmount, err := amount.FromString(req.GetAmount())
	if err != nil {
		replyError = "invalid amount"
		return
	}

	// enqueue
	if dups, ok := n.api.EnqueueSending(types.SendingNats, req.GetId(), req.GetService(), "", reqAddr, reqAmount, reqToken, req.GetIgnoreApprovement()); !ok {
		if dups {
			replyError = "request with the same ID registered"
		} else {
			replyError = "internal failure"
		}
		return
	}
}

// subApproveRequest listens for a new approvement requests until connection draining
func (n *Nats) subApproveRequest(m *gonats.Msg) {
	nc := n.natsConnection

	// metrics
	if n.metrics != nil {
		defer func(t time.Time, method string) {
			n.metrics.RequestDuration.WithLabelValues("approve").Observe(time.Since(t).Seconds())
		}(time.Now(), m.Subject)
	}

	// parse
	req := senderNats.Approve{}
	if err := proto.Unmarshal(m.Data, &req); err != nil {
		n.logger.WithError(err).Error("Failed to unmarshal request")
		return
	}

	n.logger.WithField("data", req.String()).Debug("Got approvement request")

	// reply
	var replyError string
	defer func() {
		rep := senderNats.ApproveReply{
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
	if !model.ServiceNameRex.MatchString(req.GetService()) {
		replyError = "invalid service name"
		return
	}

	// check req id
	if !model.RequestIDRex.MatchString(req.GetId()) {
		replyError = "invalid request ID"
		return
	}

	// parse wallet address
	reqAddr, err := mint.ParsePublicKey(req.GetPublicKey())
	if err != nil {
		replyError = "invalid public key"
		return
	}

	// enqueue
	if dups, ok := n.api.EnqueueApprovement(types.SendingNats, req.GetId(), req.GetService(), "", reqAddr); !ok {
		if dups {
			replyError = "request with the same ID registered"
		} else {
			replyError = "internal failure"
		}
		return
	}
}
