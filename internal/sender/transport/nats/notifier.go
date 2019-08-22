package nats

import (
	"time"

	proto "github.com/golang/protobuf/proto"
	senderNatsProto "github.com/void616/gm-mint-sender/pkg/sender/nats/sender"
	sumuslib "github.com/void616/gm-sumuslib"
)

// PublishSentEvent sends a "sent" event
func (s *Service) PublishSentEvent(id string, success bool, errorDesc string, tx sumuslib.Digest) bool {
	// metrics
	mt := time.Now()

	reqModel := senderNatsProto.SentEvent{
		Id:          id,
		Success:     success,
		Error:       errorDesc,
		Transaction: sumuslib.Pack58(tx[:]),
	}

	req, err := proto.Marshal(&reqModel)
	if err != nil {
		s.logger.WithError(err).Error("Failed to marshal")
		return false
	}

	msg, err := s.natsConnection.Request(s.subjPrefix+senderNatsProto.SubjectSent, req, time.Second*5)
	if err != nil {
		s.logger.WithError(err).Error("Failed to send request")
		return false
	}

	repModel := senderNatsProto.SentEventReply{}
	if err := proto.Unmarshal(msg.Data, &repModel); err != nil {
		s.logger.WithError(err).Error("Failed to unmarshal")
		return false
	}

	if !repModel.GetSuccess() {
		s.logger.Errorf("Unsuccessful reply: %v", repModel.GetError())
		return false
	}

	// metrics
	if s.mtxTaskDuration != nil {
		s.mtxTaskDuration.WithLabelValues("nats_notifier").Observe(time.Since(mt).Seconds())
	}

	return true
}
