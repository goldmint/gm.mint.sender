package senderservice

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"github.com/void616/gm-mint-sender/internal/sender/db"
)

// Service provides methods to enqueue sending
type Service struct {
	logger *logrus.Entry
	dao    db.DAO

	mtxMethodDuration *prometheus.SummaryVec
}

// New Service instance
func New(
	dao db.DAO,
	mtxMethodDuration *prometheus.SummaryVec,
	logger *logrus.Entry,
) (*Service, error) {
	f := &Service{
		logger:            logger,
		dao:               dao,
		mtxMethodDuration: mtxMethodDuration,
	}
	return f, nil
}
