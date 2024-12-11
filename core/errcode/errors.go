package errcode

import (
	"github.com/pkg/errors"
)

var (
	ErrPublisherNotSet = errors.New("publisher not found")
	ErrStreamClosed    = errors.New("stream closed")
	//ErrSubscriberNotFound = errors.New("subscriber not found")
)
