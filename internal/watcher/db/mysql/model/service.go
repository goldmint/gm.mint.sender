package model

import (
	"github.com/void616/gm-mint-sender/internal/watcher/db/types"
)

// Service model
type Service struct {
	Base
	ID          uint64 `gorm:"PRIMARY_KEY;AUTO_INCREMENT:true;NOT NULL"`
	Name        string `gorm:"SIZE:64;NOT NULL"`
	Transport   uint8  `gorm:"NOT NULL"`
	CallbackURL string `gorm:"SIZE:256;NOT NULL"`
}

// MapFrom mapping
func (s *Service) MapFrom(t *types.Service) error {
	s.ID = t.ID
	s.Name = LimitStringField(t.Name, 64)
	s.Transport = uint8(t.Transport)
	s.CallbackURL = LimitStringField(t.CallbackURL, 256)
	return nil
}

// MapTo mapping
func (s *Service) MapTo() (*types.Service, error) {
	return &types.Service{
		ID:          s.ID,
		Name:        s.Name,
		Transport:   types.ServiceTransport(s.Transport),
		CallbackURL: s.CallbackURL,
	}, nil
}
