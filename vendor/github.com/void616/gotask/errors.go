package gotask

import "errors"

var (
	// ErrInvalidRoutine means routine has invalid type
	ErrInvalidRoutine = errors.New("invalid routine")
	// ErrInvalidState means a task/group state is invalid for current operation (broken flow)
	ErrInvalidState = errors.New("invalid state")
)
