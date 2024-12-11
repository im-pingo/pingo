package stream

import (
	"context"
	"io"
	"testing"

	"github.com/pingostack/pingos/pkg/avframe"
	"github.com/stretchr/testify/assert"
)

// MockPublisher is a mock implementation of a publisher for testing purposes
type MockPublisher struct {
	message string
}

func (mp *MockPublisher) Publish(msg string) error {
	mp.message = msg
	return nil
}

func (mp *MockPublisher) ID() string {
	return "mockPublisher"
}

func (mp *MockPublisher) Format() avframe.FmtType {
	return avframe.FmtType(0) // Return a mock format type
}

func (mp *MockPublisher) Metadata() avframe.Metadata {
	return avframe.Metadata{} // Return a mock metadata
}

func (mp *MockPublisher) Close() error {
	return nil // Simulate closing the publisher
}

func (mp *MockPublisher) Read() (*avframe.Frame, error) {
	return &avframe.Frame{}, io.EOF // Mock implementation of Read
}

func (mp *MockPublisher) ReadFrame() (*avframe.Frame, error) {
	return &avframe.Frame{}, nil // Mock implementation of Read
}

// MockSubscriber is a mock implementation of a subscriber for testing purposes
type MockSubscriber struct {
	onMessage func(msg string)
}

func (ms *MockSubscriber) Subscribe(onMessage func(msg string)) {
	ms.onMessage = onMessage
}

func (ms *MockSubscriber) ID() string {
	return "mockSubscriber"
}

func (ms *MockSubscriber) Format() avframe.FmtType {
	return avframe.FmtType(0) // Return a mock format type
}

func (ms *MockSubscriber) AudioCodecSupported() []avframe.CodecType {
	return []avframe.CodecType{} // Return a mock codec type
}

func (ms *MockSubscriber) VideoCodecSupported() []avframe.CodecType {
	return []avframe.CodecType{} // Return a mock codec type
}

func (ms *MockSubscriber) SetProcessor(processor avframe.Processor) {
	// Mock implementation
}

func (ms *MockSubscriber) Close() error {
	return nil // Simulate closing the subscriber
}

func (ms *MockSubscriber) Read() (*avframe.Frame, error) {
	return &avframe.Frame{}, io.EOF // Mock implementation of Read
}

func (ms *MockSubscriber) Write(f *avframe.Frame) error {
	return nil // Mock implementation of Write
}

func (ms *MockSubscriber) WriteFrame(frame *avframe.Frame) error {
	return nil // Mock implementation of Write
}

func TestStreamPublisherOperations(t *testing.T) {
	// Initialize a new Stream instance with a context and ID
	stream := NewStream(context.Background(), "streamID")

	// Test setting a publisher
	t.Run("set publisher", func(t *testing.T) {
		publisher := &MockPublisher{} // Mock publisher implementation
		err := stream.Publish(publisher)
		assert.NoError(t, err, "Expected no error when setting a publisher")
	})

	// Test adding a subscriber
	t.Run("add subscriber", func(t *testing.T) {
		received := make(chan string)
		subscriber := &MockSubscriber{onMessage: func(msg string) {
			received <- msg
		}}
		_, err := stream.Subscribe(subscriber, nil)
		assert.NoError(t, err, "Expected no error when adding a subscriber")
		msg := "Hello, Subscriber!"
		err = stream.Publish(&MockPublisher{message: msg})
		assert.NoError(t, err, "Expected no error when publishing a message")
		assert.Equal(t, msg, <-received, "Expected the received message to match the published message")
	})

	// Test error handling during adding a subscriber
	t.Run("add subscriber error", func(t *testing.T) {
		stream.Close() // Assuming Close method is available
		_, err := stream.Subscribe(&MockSubscriber{}, nil)
		assert.Error(t, err, "Expected an error when adding a subscriber to a closed stream")
	})
}
