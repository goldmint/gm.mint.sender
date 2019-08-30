package nats

import (
	"time"
	"unicode/utf8"

	proto "github.com/golang/protobuf/proto"
	gonats "github.com/nats-io/go-nats"
	walletNats "github.com/void616/gm-mint-sender/pkg/watcher/nats/wallet"
	sumuslib "github.com/void616/gm-sumuslib"
)

// addRemoveWallet processes Nats request to add/remove a wallet
func (s *Service) addRemoveWallet(m *gonats.Msg) {
	nc := s.natsConnection

	// metrics
	if s.mtxRequestDuration != nil {
		defer func(t time.Time, method string) {
			s.mtxRequestDuration.WithLabelValues(method).Observe(time.Since(t).Seconds())
		}(time.Now(), m.Subject)
	}

	// parse
	req := walletNats.AddRemoveRequest{}
	if err := proto.Unmarshal(m.Data, &req); err != nil {
		s.logger.WithError(err).Error("Failed to unmarshal request")
		return
	}

	s.logger.WithField("data", req.String()).Trace("Got wallet request")

	// reply
	var replyError string
	defer func() {
		rep := walletNats.AddRemoveReply{
			Success: replyError == "",
			Error:   replyError,
		}
		if b, err := proto.Marshal(&rep); err != nil {
			s.logger.WithError(err).Error("Failed to marshal reply")
		} else {
			if err := nc.Publish(m.Reply, b); err != nil {
				s.logger.WithError(err).Error("Failed to publish reply")
			}
		}
	}()

	// check req service
	if x := utf8.RuneCountInString(req.GetService()); x < 1 || x > 64 || !serviceNameRex.MatchString(req.GetService()) {
		replyError = "Invalid service name"
		return
	}

	// unpack base58
	pubs := make([]sumuslib.PublicKey, 0)
	for _, p := range req.GetPublicKey() {
		pub, err := sumuslib.ParsePublicKey(p)
		if err != nil {
			replyError = "One or more invalid Base58 public keys"
			return
		}
		pubs = append(pubs, pub)
	}
	if req.GetAdd() {
		if !s.walletService.AddWallet(req.GetService(), pubs...) {
			replyError = "Internal failure"
			return
		}
	} else {
		if !s.walletService.RemoveWallet(req.GetService(), pubs...) {
			replyError = "Internal failure"
			return
		}
	}
}
