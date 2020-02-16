package nats

import (
	"time"

	proto "github.com/golang/protobuf/proto"
	gonats "github.com/nats-io/nats.go"
	mint "github.com/void616/gm.mint"
	"github.com/void616/gm.mint.sender/internal/watcher/api/model"
	"github.com/void616/gm.mint.sender/internal/watcher/db/types"
	walletNats "github.com/void616/gm.mint.sender/pkg/watcher/nats"
)

// subAddRemoveWallet processes Nats request to add/remove a wallet
func (n *Nats) subAddRemoveWallet(m *gonats.Msg) {
	nc := n.natsConnection

	// metrics
	if n.metrics != nil {
		defer func(t time.Time) {
			n.metrics.RequestDuration.WithLabelValues("add_wallet").Observe(time.Since(t).Seconds())
		}(time.Now())
	}

	// parse
	req := walletNats.AddRemove{}
	if err := proto.Unmarshal(m.Data, &req); err != nil {
		n.logger.WithError(err).Error("Failed to unmarshal request")
		return
	}

	n.logger.WithField("data", len(req.GetPublicKey())).Debug("Got wallet request")

	// reply
	var replyError string
	defer func() {
		rep := walletNats.AddRemoveReply{
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

	// unpack base58
	pubs := make([]mint.PublicKey, 0)
	for _, p := range req.GetPublicKey() {
		pub, err := mint.ParsePublicKey(p)
		if err != nil {
			replyError = "one or more invalid Base58 public keys"
			return
		}
		pubs = append(pubs, pub)
	}
	if req.GetAdd() {
		if !n.api.AddWallet(types.ServiceNats, req.GetService(), "", pubs...) {
			replyError = "internal failure"
			return
		}
	} else {
		if !n.api.RemoveWallet(req.GetService(), pubs...) {
			replyError = "internal failure"
			return
		}
	}
}
