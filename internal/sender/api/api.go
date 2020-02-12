package api

import (
	"github.com/sirupsen/logrus"
	"github.com/void616/gm.mint.sender/internal/sender/db"
)

// API provides methods to enqueue sending
type API struct {
	logger *logrus.Entry
	dao    db.DAO
}

// New instance
func New(
	dao db.DAO,
	logger *logrus.Entry,
) (*API, error) {
	f := &API{
		logger: logger,
		dao:    dao,
	}
	return f, nil
}
