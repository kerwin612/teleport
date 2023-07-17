package model

import (
	"context"
	"strings"
	"sync"

	"github.com/gravitational/trace"
	"github.com/sashabaranov/go-openai"
	"github.com/tiktoken-go/tokenizer/codec"
)

var defaultTokenizer = codec.NewCl100kBase()

type TokenCount struct {
	Prompt     TokenCounters
	Completion TokenCounters
}

func (tc *TokenCount) AddPromptCounter(prompt TokenCounter) {
	if prompt != nil {
		tc.Prompt = append(tc.Prompt, prompt)
	}
}

func (tc *TokenCount) AddCompletionCounter(completion TokenCounter) {
	if completion != nil {
		tc.Completion = append(tc.Completion, completion)
	}
}

func (tc *TokenCount) CountAll(ctx context.Context) (int, int, error) {
	prompt, err := tc.Prompt.CountAll(ctx)
	if err != nil {
		return 0, 0, trace.Wrap(err)
	}
	completion, err := tc.Completion.CountAll(ctx)
	if err != nil {
		return 0, 0, trace.Wrap(err)
	}
	return prompt, completion, nil
}

func NewTokenCount() *TokenCount {
	return &TokenCount{
		Prompt:     TokenCounters{},
		Completion: TokenCounters{},
	}
}

type TokenCounter interface {
	TokenCount(ctx context.Context) (int, error)
}

type TokenCounters []TokenCounter

func (tc TokenCounters) CountAll(ctx context.Context) (int, error) {
	var total int
	for _, counter := range tc {
		count, err := counter.TokenCount(ctx)
		if err != nil {
			return 0, trace.Wrap(err)
		}
		total += count
	}
	return total, nil
}

type PromptTokenCounter struct {
	count int
}

func (tc *PromptTokenCounter) TokenCount(_ context.Context) (int, error) {
	return tc.count, nil
}

func NewPromptTokenCounter(prompt []openai.ChatCompletionMessage) (*PromptTokenCounter, error) {
	var promptCount int
	for _, message := range prompt {
		promptTokens, _, err := defaultTokenizer.Encode(message.Content)
		if err != nil {
			return nil, trace.Wrap(err)
		}

		promptCount = promptCount + perMessage + perRole + len(promptTokens)
	}

	return &PromptTokenCounter{count: promptCount}, nil
}

type MessageTokenCounter struct {
	count int
}

func (tc *MessageTokenCounter) TokenCount(_ context.Context) (int, error) {
	return tc.count, nil
}

func NewMessageTokenCounter(completion string) (*MessageTokenCounter, error) {
	completionTokens, _, err := defaultTokenizer.Encode(completion)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	completionCount := perRequest + len(completionTokens)

	return &MessageTokenCounter{count: completionCount}, nil
}

type StreamingMessageTokenCounter struct {
	completion *strings.Builder
	count      int

	// mutex protects all fields of the StreamingMessageTokenCounter, it must be
	// acquired before any read or write operation.
	// When the count is not finished yet, mutex must be released before waiting.
	mutex    sync.Mutex
	finished bool
	// waiting is a list of channels of callers wanting to be notified by the
	// end of the token count. The channels MUST be buffered and have at least 1
	// available capacity to avoid blocking Finish()
	waiting []chan struct{}
}

func (tc *StreamingMessageTokenCounter) TokenCount(ctx context.Context) (int, error) {
	// If the count is already finished, we return the values
	tc.mutex.Lock()
	if tc.finished {
		defer tc.mutex.Unlock()
		return tc.count, nil
	}

	// Else we register and wait for the count to be finished
	// The channel is buffered with a capacity of one, to avoid the sender to
	// block if the listener gave up because of a cancelled context.
	waitC := make(chan struct{}, 1)
	tc.waiting = append(tc.waiting, waitC)
	tc.mutex.Unlock()

	select {
	case <-waitC:
		return tc.count, nil
	case <-ctx.Done():
		return 0, trace.Wrap(ctx.Err())
	}
}

func (tc *StreamingMessageTokenCounter) Finish() error {
	tc.mutex.Lock()
	defer tc.mutex.Unlock()

	// We compute the final count
	completionTokens, _, err := defaultTokenizer.Encode(tc.completion.String())
	if err != nil {
		return trace.Wrap(err)
	}
	tc.count = perRequest + len(completionTokens)
	tc.finished = true

	// And notify everyone that the job is done
	for _, waiter := range tc.waiting {
		waiter <- struct{}{}
	}
	return nil
}

func (tc *StreamingMessageTokenCounter) Add(delta string) error {
	tc.mutex.Lock()
	defer tc.mutex.Unlock()

	if tc.finished {
		return trace.Errorf("Count is already finished, cannot add more content")
	}
	tc.completion.WriteString(delta)
	return nil
}

func NewStreamingMessageTokenCounter(completionStart string) *StreamingMessageTokenCounter {
	completion := &strings.Builder{}
	completion.WriteString(completionStart)

	return &StreamingMessageTokenCounter{
		completion: completion,
		count:      0,
		mutex:      sync.Mutex{},
		finished:   false,
		waiting:    nil,
	}
}
