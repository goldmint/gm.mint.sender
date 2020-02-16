package api

import (
	"github.com/sirupsen/logrus"
	"github.com/void616/gm.mint.sender/internal/mint/rpcpool"
	"github.com/void616/gm.mint.sender/internal/sender/db"
)

// API provides methods to enqueue sending
type API struct {
	logger *logrus.Entry
	dao    db.DAO
	pool   *rpcpool.Pool
}

// New instance
func New(
	dao db.DAO,
	pool *rpcpool.Pool,
	logger *logrus.Entry,
) (*API, error) {
	f := &API{
		logger: logger,
		dao:    dao,
		pool:   pool,
	}
	return f, nil
}
