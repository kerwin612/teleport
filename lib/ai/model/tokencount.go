/*
Copyright 2023 Gravitational, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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

// TokenCount holds TokenCounters for both Prompt and Completion tokens.
// As the agent performs multiple calls to the model, each call creates its own
// prompt and completion TokenCounter.
//
// Prompt TokenCounters can be created before doing the call as we know the
// full prompt and can tokenize it. This is the PromptTokenCounter purpose.
//
// Completion TokenCounters can be created after receiving the model response.
// Depending on the response type, we might have the full result already or get
// a stream that will provide the completion result in the future. For the latter,
// the token count will be evaluated lazily and asynchronously.
// SynchronousTokenCounter count tokens synchronously, while
// AsynchornousTokenCounter supports the streaming use-cases.
type TokenCount struct {
	Prompt     TokenCounters
	Completion TokenCounters
}

// AddPromptCounter adds a TokenCounter to the Prompt list.
func (tc *TokenCount) AddPromptCounter(prompt TokenCounter) {
	if prompt != nil {
		tc.Prompt = append(tc.Prompt, prompt)
	}
}

// AddCompletionCounter adds a TokenCounter to the Completion list.
func (tc *TokenCount) AddCompletionCounter(completion TokenCounter) {
	if completion != nil {
		tc.Completion = append(tc.Completion, completion)
	}
}

// CountAll iterates over all counters and returns how many prompt and
// completion tokens were used. As completion token counting can require waiting
// for a response to be streamed, the caller should pass a context and use it to
// implement some kind of deadline to avoid hanging infinitely if something goes
// wrong (e.g. use `context.WithTimeout()`).
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

// NewTokenCount initializes a new TokenCount struct.
func NewTokenCount() *TokenCount {
	return &TokenCount{
		Prompt:     TokenCounters{},
		Completion: TokenCounters{},
	}
}

// TokenCounter is an interface for all token counters, regardless of the kind
// of token they count (prompt/completion) or the tokenizer used.
// TokenCount must be idempotent.
type TokenCounter interface {
	TokenCount(ctx context.Context) (int, error)
}

// TokenCounters is a list of TokenCounter and offers function to iterate over
// all counters and compute the total.
type TokenCounters []TokenCounter

// CountAll iterates over a list of TokenCounter and returns the sum of the
// results of all counters. As the counting process might be blocking/take some
// time, the caller should set a Deadline on the context.
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

// PromptTokenCounter implements the TokenCounter interface and counts tokens
// for a given prompt. Tokenization happens on initialization.
type PromptTokenCounter struct {
	count int
}

// TokenCount implements the TokenCounter interface.
func (tc *PromptTokenCounter) TokenCount(_ context.Context) (int, error) {
	return tc.count, nil
}

// NewPromptTokenCounter takes a list of openai.ChatCompletionMessage and
// computes how many tokens are used by sending those messages to the model.
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

// SynchronousTokenCounter counts completion tokens that have been used by a
// completion request. This can be used only if the completion request is over
// and we know the full result. Tokenization happens on initialization.
type SynchronousTokenCounter struct {
	count int
}

// TokenCount implements the TokenCounter interface.
func (tc *SynchronousTokenCounter) TokenCount(_ context.Context) (int, error) {
	return tc.count, nil
}

// NewSynchronousTokenCounter takes the completion request output and
// computes how many tokens were used by the model to generate this result.
func NewSynchronousTokenCounter(completion string) (*SynchronousTokenCounter, error) {
	completionTokens, _, err := defaultTokenizer.Encode(completion)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	completionCount := perRequest + len(completionTokens)

	return &SynchronousTokenCounter{count: completionCount}, nil
}

// AsynchornousTokenCounter counts completion tokens that are used by a
// streamed completion request. When creating a AsynchornousTokenCounter,
// the streaming might not be finished, and we can't evaluate how many tokens
// will be used. In this case, the streaming routine must add streamed
// completion result with the Add() method and call Finish() once the
// completion is finished. TokenCount() will hang until either Finish() is
// called or the context is Done.
type AsynchornousTokenCounter struct {
	completion *strings.Builder
	count      int

	// mutex protects all fields of the AsynchornousTokenCounter, it must be
	// acquired before any read or write operation.
	// When the count is not finished yet, mutex must be released before waiting.
	mutex sync.Mutex
	// finished tells if the count is finished or not.
	// if the count is already finished, Add() and Finish() will fail.
	finished bool
	// waiting is a list of channels of callers wanting to be notified by the
	// end of the token count. The channels MUST be buffered and have at least 1
	// available capacity to avoid blocking Finish()
	waiting []chan struct{}
}

// TokenCount implements the TokenCounter interface.
// If the count is already finished, it immediately returns with the count.
// If the count is not yet finished, it registers a channel to be notified when
// Finish() is called and waits until either Finish() is called or the context
// is done.
func (tc *AsynchornousTokenCounter) TokenCount(ctx context.Context) (int, error) {
	// If the count is already finished, we return the values
	tc.mutex.Lock()
	if tc.finished {
		defer tc.mutex.Unlock()
		return tc.count, nil
	}

	// Else we register and wait for the count to be finished
	// The channel is buffered with a capacity of one, to avoid the sender to
	// block if the listener gave up because of a canceled context.
	waitC := make(chan struct{}, 1)
	tc.waiting = append(tc.waiting, waitC)
	tc.mutex.Unlock()

	select {
	case <-waitC:
		// No need to acquire lock here because we know the count already
		// happened and the values cannot change anymore because tc.finished is
		// set.
		return tc.count, nil
	case <-ctx.Done():
		return 0, trace.Wrap(ctx.Err())
	}
}

// Finish marks the token counting process as over. Once finished is called, the
// token count is computed and the count cannot change anymore. Finish also
// notifies all waiters that the count is now available.
func (tc *AsynchornousTokenCounter) Finish() error {
	tc.mutex.Lock()
	defer tc.mutex.Unlock()
	if tc.finished {
		return trace.Errorf("Already finished counting tokens")
	}

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

// Add a streamed completion response bit to the to-be-counted result.
func (tc *AsynchornousTokenCounter) Add(delta string) error {
	tc.mutex.Lock()
	defer tc.mutex.Unlock()

	if tc.finished {
		return trace.Errorf("Count is already finished, cannot add more content")
	}
	tc.completion.WriteString(delta)
	return nil
}

// NewAsynchronousTokenCounter takes the partial completion request output
// and creates a token counter that can be already returned even if not all
// the content has been streamed yet. Streamed content can be added a posteriori
// with Add(). Once all the content is streamed, Finish() must be called.
func NewAsynchronousTokenCounter(completionStart string) *AsynchornousTokenCounter {
	completion := &strings.Builder{}
	completion.WriteString(completionStart)

	return &AsynchornousTokenCounter{
		completion: completion,
		count:      0,
		mutex:      sync.Mutex{},
		finished:   false,
		waiting:    nil,
	}
}
