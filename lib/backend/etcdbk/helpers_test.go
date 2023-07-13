package etcdbk

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRoundRobinConcurrent(t *testing.T) {
	t.Parallel()

	const workers = 100
	const rounds = 100

	rr := newRoundRobin([]bool{true, false})

	var tct atomic.Uint64
	var fct atomic.Uint64

	var wg sync.WaitGroup

	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for r := 0; r < rounds; r++ {
				if rr.Next() {
					tct.Add(1)
				} else {
					fct.Add(1)
				}
			}
		}()
	}

	wg.Wait()

	require.Equal(t, workers*rounds, int(tct.Load()+fct.Load()))
	require.InDelta(t, tct.Load(), fct.Load(), 1.0)
}

func TestRoundRobinSequential(t *testing.T) {
	t.Parallel()
	tts := []struct {
		desc   string
		items  []string
		expect []string
	}{
		{
			desc:  "single-item",
			items: []string{"foo"},
			expect: []string{
				"foo",
				"foo",
				"foo",
			},
		},
		{
			desc: "multi-item",
			items: []string{
				"foo",
				"bar",
				"bin",
				"baz",
			},
			expect: []string{
				"foo",
				"bar",
				"bin",
				"baz",
				"foo",
				"bar",
				"bin",
				"baz",
				"foo",
				"bar",
				"bin",
				"baz",
			},
		},
	}
	for _, tt := range tts {
		t.Run(tt.desc, func(t *testing.T) {
			rr := newRoundRobin(tt.items)
			for _, exp := range tt.expect {
				require.Equal(t, exp, rr.Next())
			}
		})
	}
}
