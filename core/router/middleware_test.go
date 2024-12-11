package router

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/pingostack/pingos/core/stream"
	"github.com/stretchr/testify/assert"
)

func TestBuildStreamMiddleware(t *testing.T) {
	// Reset the stream chain before each test
	streamChain = &MiddlewareChain[*stream.Stream]{}

	t.Run("single middleware", func(t *testing.T) {
		executed := false
		middleware := func(ctx context.Context, s *stream.Stream, stage Stage) error {
			executed = true
			return nil
		}

		RegisterStreamMiddleware(middleware)
		final := func(ctx context.Context, s *stream.Stream, stage Stage) error {
			return nil
		}

		chain := BuildStreamMiddleware(context.Background(), final)
		err := chain(context.Background(), &stream.Stream{}, StageStart)

		assert.NoError(t, err)
		assert.True(t, executed)
	})

	t.Run("multiple middlewares execution order", func(t *testing.T) {
		streamChain = &MiddlewareChain[*stream.Stream]{} // Reset chain
		executionOrder := []int{}

		middleware1 := func(ctx context.Context, s *stream.Stream, stage Stage) error {
			executionOrder = append(executionOrder, 1)
			return nil
		}
		middleware2 := func(ctx context.Context, s *stream.Stream, stage Stage) error {
			executionOrder = append(executionOrder, 2)
			return nil
		}
		middleware3 := func(ctx context.Context, s *stream.Stream, stage Stage) error {
			executionOrder = append(executionOrder, 3)
			return nil
		}

		RegisterStreamMiddleware(middleware1)
		RegisterStreamMiddleware(middleware2)
		RegisterStreamMiddleware(middleware3)

		final := func(ctx context.Context, s *stream.Stream, stage Stage) error {
			executionOrder = append(executionOrder, 4)
			return nil
		}

		chain := BuildStreamMiddleware(context.Background(), final)
		err := chain(context.Background(), &stream.Stream{}, StageStart)

		assert.NoError(t, err)
		assert.Equal(t, []int{1, 2, 3, 4}, executionOrder)
	})

	t.Run("error propagation", func(t *testing.T) {
		streamChain = &MiddlewareChain[*stream.Stream]{} // Reset chain
		expectedError := errors.New("middleware error")
		executed := false

		middleware1 := func(ctx context.Context, s *stream.Stream, stage Stage) error {
			return expectedError
		}
		middleware2 := func(ctx context.Context, s *stream.Stream, stage Stage) error {
			executed = true // This should not execute due to earlier error
			return nil
		}

		RegisterStreamMiddleware(middleware1)
		RegisterStreamMiddleware(middleware2)

		final := func(ctx context.Context, s *stream.Stream, stage Stage) error {
			executed = true // This should not execute due to earlier error
			return nil
		}

		chain := BuildStreamMiddleware(context.Background(), final)
		err := chain(context.Background(), &stream.Stream{}, StageStart)

		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
		assert.False(t, executed)
	})

	t.Run("context propagation", func(t *testing.T) {
		streamChain = &MiddlewareChain[*stream.Stream]{} // Reset chain
		key := struct{}{}
		value := "test-value"
		var receivedValue interface{}

		middleware := func(ctx context.Context, s *stream.Stream, stage Stage) error {
			receivedValue = ctx.Value(key)
			return nil
		}

		RegisterStreamMiddleware(middleware)
		final := func(ctx context.Context, s *stream.Stream, stage Stage) error {
			return nil
		}

		ctx := context.WithValue(context.Background(), key, value)
		chain := BuildStreamMiddleware(ctx, final)
		err := chain(ctx, &stream.Stream{}, StageStart)

		assert.NoError(t, err)
		assert.Equal(t, value, receivedValue)
	})

	t.Run("execution chain verification", func(t *testing.T) {
		streamChain = &MiddlewareChain[*stream.Stream]{} // Reset chain
		executionLog := []string{}

		// First middleware
		middleware1 := func(ctx context.Context, s *stream.Stream, stage Stage) error {
			executionLog = append(executionLog, fmt.Sprintf("middleware1 start - stage: %v", stage))
			return nil
		}

		// Second middleware with both pre and post logging
		middleware2 := func(ctx context.Context, s *stream.Stream, stage Stage) error {
			executionLog = append(executionLog, fmt.Sprintf("middleware2 start - stage: %v", stage))
			return nil
		}

		// Third middleware that verifies context
		middleware3 := func(ctx context.Context, s *stream.Stream, stage Stage) error {
			executionLog = append(executionLog, fmt.Sprintf("middleware3 start - stage: %v", stage))
			return nil
		}

		// Register all middlewares
		RegisterStreamMiddleware(middleware1)
		RegisterStreamMiddleware(middleware2)
		RegisterStreamMiddleware(middleware3)

		// Final handler
		final := func(ctx context.Context, s *stream.Stream, stage Stage) error {
			executionLog = append(executionLog, fmt.Sprintf("final handler - stage: %v", stage))
			return nil
		}

		// Build and execute the chain for both stages
		chain := BuildStreamMiddleware(context.Background(), final)

		// Test StageStart
		err := chain(context.Background(), &stream.Stream{}, StageStart)
		assert.NoError(t, err)

		// Test StageEnd
		err = chain(context.Background(), &stream.Stream{}, StageEnd)
		assert.NoError(t, err)

		// Print execution log for verification
		t.Log("Execution Chain Log:")
		for i, log := range executionLog {
			t.Logf("%d: %s", i+1, log)
		}

		// Verify execution order for StageStart
		assert.Equal(t, "middleware1 start - stage: start", executionLog[0])
		assert.Equal(t, "middleware2 start - stage: start", executionLog[1])
		assert.Equal(t, "middleware3 start - stage: start", executionLog[2])
		assert.Equal(t, "final handler - stage: start", executionLog[3])

		// Verify execution order for StageEnd
		assert.Equal(t, "middleware1 start - stage: end", executionLog[4])
		assert.Equal(t, "middleware2 start - stage: end", executionLog[5])
		assert.Equal(t, "middleware3 start - stage: end", executionLog[6])
		assert.Equal(t, "final handler - stage: end", executionLog[7])

		// Verify total number of executions
		assert.Equal(t, 8, len(executionLog), "Expected 8 middleware executions (4 for each stage)")
	})

	t.Run("error chain verification", func(t *testing.T) {
		streamChain = &MiddlewareChain[*stream.Stream]{} // Reset chain
		executionLog := []string{}
		expectedError := errors.New("middleware2 error")

		middleware1 := func(ctx context.Context, s *stream.Stream, stage Stage) error {
			executionLog = append(executionLog, "middleware1 executed")
			return nil
		}

		middleware2 := func(ctx context.Context, s *stream.Stream, stage Stage) error {
			executionLog = append(executionLog, "middleware2 executed")
			return expectedError
		}

		middleware3 := func(ctx context.Context, s *stream.Stream, stage Stage) error {
			executionLog = append(executionLog, "middleware3 executed")
			return nil
		}

		RegisterStreamMiddleware(middleware1)
		RegisterStreamMiddleware(middleware2)
		RegisterStreamMiddleware(middleware3)

		final := func(ctx context.Context, s *stream.Stream, stage Stage) error {
			executionLog = append(executionLog, "final executed")
			return nil
		}

		chain := BuildStreamMiddleware(context.Background(), final)
		err := chain(context.Background(), &stream.Stream{}, StageStart)

		// Print execution log for verification
		t.Log("Error Chain Log:")
		for i, log := range executionLog {
			t.Logf("%d: %s", i+1, log)
		}

		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
		assert.Equal(t, 2, len(executionLog), "Should only execute up to error middleware")
		assert.Equal(t, []string{
			"middleware1 executed",
			"middleware2 executed",
		}, executionLog)
	})
}
