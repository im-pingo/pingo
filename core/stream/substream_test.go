package stream

import (
	"context"
	"testing"
	"time"
)

func TestUpdateDeadline(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sub := &subStream{
		ctx: ctx,
	}

	// Update the deadline to 1 second from now
	sub.SetDeadline(time.Now().Add(1 * time.Second))

	// Wait for the timer to expire
	time.Sleep(2 * time.Second)

	// Check if the substream is closed
	// You might need to implement a way to check if the substream is closed
	// For example, by adding a closed field in the subStream struct
	// assert.True(t, sub.isClosed, "Expected substream to be closed")
}
