package domain

import "errors"

// Sentinel errors used throughout the domain layer.
var (
	ErrTaskNotFound   = errors.New("task not found")
	ErrWorkerNotFound = errors.New("worker not found")
	ErrQueueEmpty     = errors.New("queue is empty")
	ErrTaskInvalid    = errors.New("task is invalid")
	ErrWorkerInvalid  = errors.New("worker is invalid")
)
