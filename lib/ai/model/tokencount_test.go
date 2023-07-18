package model

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testCompletionStart       = "This is the beginning of the response."
	testCompletionEnd         = " And this is the end."
	testCompletionStartTokens = 8 // 1 token per word + 1 for the dot
	testCompletionEndTokens   = 6 // 1 token per word + 1 for the dot
	testCompletionTokens      = testCompletionStartTokens + testCompletionEndTokens
)

// This test checks that TokenCount() returns instantly on an already
// finished count.
func TestAsynchronousTokenCount_Finished(t *testing.T) {
	t.Parallel()
	// Test setup: we create an already finished counter
	tc := NewAsynchronousTokenCounter(testCompletionStart)
	err := tc.Add(testCompletionEnd)
	require.NoError(t, err)
	require.NoError(t, tc.Finish())

	ctx := context.Background()
	result, err := tc.TokenCount(ctx)
	require.NoError(t, err)
	require.Equal(t, testCompletionTokens+perRequest, result)
}

// This test checks that TokenCount() waits properly on a count still in
// progress. It checks the scenario in which the context gets cancelled first
func TestAsynchronousTokenCount_WaitingCancel(t *testing.T) {
	t.Parallel()

	// Test setup: we create an already finished counter
	tc := NewAsynchronousTokenCounter(testCompletionStart)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	// We expect the context to timeout
	result, err := tc.TokenCount(ctx)
	require.Error(t, err)
	require.Empty(t, result)

}

// This test checks that TokenCount() waits properly on a count still in
// progress. It checks the scenario in which a routine is waiting for the result
// and the count is finished before the context expires. The waiting routine
// be unblocked and get the correct result.
func TestAsynchronousTokenCount_WaitingFinish(t *testing.T) {
	t.Parallel()

	// Test setup: we create an already finished counter
	tc := NewAsynchronousTokenCounter(testCompletionStart)
	ctx := context.Background()

	finished := make(chan struct{})

	// We spin a routine that waits for the token count and will need to be
	// notified
	go func() {
		result, err := tc.TokenCount(ctx)
		assert.NoError(t, err)
		assert.Equal(t, testCompletionTokens+perRequest, result)
		finished <- struct{}{}
	}()

	// We make sure the routine is stuck waiting
	require.Eventually(t, func() bool {
		return len(tc.waiting) > 0
	}, 200*time.Millisecond, 20*time.Millisecond)

	// Then we add more content to the streamed completion
	require.NoError(t, tc.Add(testCompletionEnd))

	// We mark the stream as finished. This should unblock
	require.NoError(t, tc.Finish())

	// Finally we ensure the waiting routine finished
	select {
	case <-finished:
		return
	case <-time.After(200 * time.Millisecond):
		require.Fail(t, "the routine waiting for the TokenCount() never finished")
	}
}

// This test checks that Add() properly appends content in the completion
// response.
func TestAsynchronousTokenCount_Counting(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	tests := []struct {
		name            string
		completionStart string
		completionEnd   string
		expectedTokens  int
	}{
		{
			name: "empty count",
		},
		{
			name:            "only completion start",
			completionStart: testCompletionStart,
			expectedTokens:  testCompletionStartTokens,
		},
		{
			name:           "only completion add",
			completionEnd:  testCompletionEnd,
			expectedTokens: testCompletionEndTokens,
		},
		{
			name:            "completion start and end",
			completionStart: testCompletionStart,
			completionEnd:   testCompletionEnd,
			expectedTokens:  testCompletionTokens,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Test setup
			tc := NewAsynchronousTokenCounter(tt.completionStart)
			require.NoError(t, tc.Add(tt.completionEnd))
			require.NoError(t, tc.Finish())

			// Doing the real test: asserting the count is right
			count, err := tc.TokenCount(ctx)
			require.NoError(t, err)
			require.Equal(t, tt.expectedTokens+perRequest, count)
		})
	}
}
