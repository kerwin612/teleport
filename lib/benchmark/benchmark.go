/*
Copyright 2020 Gravitational, Inc.
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

// Package benchmark package provides tools to run progressive or independent benchmarks against teleport services.
package benchmark

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/HdrHistogram/hdrhistogram-go"
	"github.com/gravitational/trace"
	"github.com/sirupsen/logrus"
	"go.uber.org/atomic"
	"golang.org/x/sync/errgroup"

	tracessh "github.com/gravitational/teleport/api/observability/tracing/ssh"
	"github.com/gravitational/teleport/api/profile"
	"github.com/gravitational/teleport/api/types"
	"github.com/gravitational/teleport/lib/client"
	"github.com/gravitational/teleport/lib/observability/tracing"
	"github.com/gravitational/teleport/lib/utils"
	"github.com/gravitational/teleport/lib/web"
)

const (
	// minValue is the min millisecond recorded for histogram
	minValue = 1
	// maxValue is the max millisecond recorded for histogram
	maxValue = 60000
	// significantFigures is the precision of the values
	significantFigures = 3
	// pauseTimeBetweenBenchmarks is the time to pause between each benchmark
	pauseTimeBetweenBenchmarks = time.Second * 5
)

// Config specifies benchmark requests to run
type Config struct {
	// Rate is requests per second origination rate
	Rate int
	// Command is a command to run
	Command []string
	// Interactive turns on interactive sessions
	Interactive bool
	// Web uses web sessions instead of ssh sessions
	Web bool
	// MinimumWindow is the min duration
	MinimumWindow time.Duration
	// MinimumMeasurements is the min amount of requests
	MinimumMeasurements int
	MaxSessions         int
}

// Result is a result of the benchmark
type Result struct {
	// RequestsOriginated is amount of requests originated
	RequestsOriginated int
	// RequestsFailed is amount of requests failed
	RequestsFailed int
	// Histogram holds the response duration values
	Histogram *hdrhistogram.Histogram
	// LastError contains last recorded error
	LastError error
	// Duration it takes for the whole benchmark to run
	Duration time.Duration
}

// Run is used to run the benchmarks, it is given a generator, command to run,
// a host, host login, and proxy. If host login or proxy is an empty string, it will
// use the default login
func Run(ctx context.Context, lg *Linear, cmd, host, login, proxy string) ([]Result, error) {
	c := strings.Split(cmd, " ")
	lg.config = &Config{Command: c}
	if err := validateConfig(lg); err != nil {
		return nil, trace.Wrap(err)
	}
	tc, err := makeTeleportClient(host, login, proxy)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	var results []Result
	sleep := false
	for {
		if sleep {
			select {
			case <-time.After(pauseTimeBetweenBenchmarks):
			case <-ctx.Done():
				return results, trace.ConnectionProblem(ctx.Err(), "context canceled or timed out")
			}
		}
		benchmarkC := lg.GetBenchmark()
		if benchmarkC == nil {
			break
		}
		result, err := benchmarkC.Benchmark(ctx, tc, nil)
		if err != nil {
			return results, trace.Wrap(err)
		}
		results = append(results, result)
		fmt.Printf("current generation requests: %v, duration: %v\n", result.RequestsOriginated, result.Duration)
		sleep = true
	}
	return results, nil
}

// ExportLatencyProfile exports the latency profile and returns the path as a string if no errors
func ExportLatencyProfile(path string, h *hdrhistogram.Histogram, ticks int32, valueScale float64) (string, error) {
	timestamp := time.Now().Format("2006-01-02_15:04:05")
	suffix := fmt.Sprintf("latency_profile_%s.txt", timestamp)
	if path != "." {
		if err := os.MkdirAll(path, 0o700); err != nil {
			return "", trace.Wrap(err)
		}
	}
	fullPath := filepath.Join(path, suffix)
	fo, err := os.Create(fullPath)
	if err != nil {
		return "", trace.Wrap(err)
	}

	if _, err := h.PercentilesPrint(fo, ticks, valueScale); err != nil {
		if err := fo.Close(); err != nil {
			logrus.WithError(err).Warn("failed to close file")
		}
		return "", trace.Wrap(err)
	}

	if err := fo.Close(); err != nil {
		return "", trace.Wrap(err)
	}
	return fo.Name(), nil
}

// Benchmark connects to remote server and executes requests in parallel according
// to benchmark spec. It returns benchmark result when completed.
// This is a blocking function that can be canceled via context argument.
func (c *Config) Benchmark(ctx context.Context, tc *client.TeleportClient, pc *web.ProxyClient) (Result, error) {
	tc.Stdout = io.Discard
	tc.Stderr = io.Discard
	tc.Stdin = &bytes.Buffer{}
	var delay time.Duration
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	resultC := make(chan *benchMeasure)

	go func() {
		interval := time.Duration(1 / float64(c.Rate) * float64(time.Second))
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		start := time.Now()
		for {
			select {
			case <-ticker.C:
				// ticker makes its first tick after the given duration, not immediately
				// this sets ResponseStart time accurately
				delay += interval
				t := start.Add(delay)
				measure := benchMeasure{
					ResponseStart: t,
					command:       c.Command,
					client:        tc,
					interactive:   c.Interactive,
					web:           c.Web,
				}
				go measure.work(ctx, resultC)
			case <-ctx.Done():
				return
			}
		}
	}()

	var result Result
	result.Histogram = hdrhistogram.New(minValue, maxValue, significantFigures)
	statusTicker := time.NewTicker(1 * time.Second)
	timeElapsed := false
	start := time.Now()
	for {
		if c.MinimumWindow <= time.Since(start) {
			timeElapsed = true
		}
		select {
		case measure := <-resultC:
			if err := result.Histogram.RecordValue(int64(measure.End.Sub(measure.ResponseStart) / time.Millisecond)); err != nil {
				logrus.WithError(err).Warn("failed to record histogram value")
			}
			result.RequestsOriginated++
			if timeElapsed && result.RequestsOriginated >= c.MinimumMeasurements {
				cancel()
			}
			if measure.Error != nil {
				result.RequestsFailed++
				result.LastError = measure.Error
			}
		case <-ctx.Done():
			result.Duration = time.Since(start)
			return result, nil
		case <-statusTicker.C:
			logrus.Infof("working... current observation count: %d", result.RequestsOriginated)
		}

	}
}

func (c *Config) Sessions(ctx context.Context, tc *client.TeleportClient, pc *web.ProxyClient) error {
	nodes, err := tc.ListAllNodes(ctx)
	if err != nil {
		return trace.Wrap(err)
	}

	proxyClient, err := tc.ConnectToProxy(ctx)
	if err != nil {
		return trace.Wrap(err)
	}
	defer proxyClient.Close()

	authClt, err := proxyClient.ConnectToRootCluster(ctx)
	if err != nil {
		return trace.Wrap(err)
	}

	netConfig, err := authClt.GetClusterNetworkingConfig(ctx)
	if err != nil {
		return trace.Wrap(err)
	}

	sessionCounter := atomic.NewInt64(0)

	cmd := strings.Join(c.Command, " ")

	spawnDuration := 750 * time.Millisecond
	var sess types.WebSession
	if c.Web {
		spawnDuration = 120 * time.Millisecond

		ws, err := pc.CurrentWebSession(ctx, authClt)
		if err != nil {
			return trace.Wrap(err)
		}
		sess = ws
	}

	spawnTicker := time.NewTicker(spawnDuration)
	// 3 seconds matches the web ui refresh rate
	statusTicker := time.NewTicker(3 * time.Second)
	defer spawnTicker.Stop()
	defer statusTicker.Stop()
	var g sync.WaitGroup

	worker := func(idx int64, node types.Server) {
		g.Add(1)
		go func() {
			defer g.Done()
			reader, writer := io.Pipe()
			defer reader.Close()
			defer writer.Close()
			log := logrus.WithFields(logrus.Fields{"index": idx, "node": node.GetHostname()})
			f := func() {
				g, gctx := errgroup.WithContext(ctx)

				if c.Web {
					g.Go(func() error {
						return pc.SSH(gctx, tc, node.GetHostname(), cmd, authClt, sess, netConfig.GetKeepAliveInterval(), func(terminal io.ReadWriteCloser) error {
							sessionCounter.Inc()

							// we cannot block here otherwise the session will never be
							// fully established. Launch another task so that we can properly
							// clean up the session once it is completed or killed
							g.Go(func() error {
								defer sessionCounter.Dec()

								// hold the session open until either of these
								// contexts is canceled
								select {
								case <-ctx.Done():
								case <-gctx.Done():
								}

								log.Debug("------ exiting")
								// send the exit command to kill the ssh session
								_, err := terminal.Write([]byte("\r\nexit\r\n"))
								return trace.Wrap(err)
							})

							select {
							case <-ctx.Done():
								return nil
							case <-gctx.Done():
								return nil
							default:
							}

							if _, err := terminal.Write([]byte(cmd + "\r\n")); err != nil {
								return trace.Wrap(err)
							}

							return nil
						})
					})
				} else {
					clt, err := tc.Clone()
					if err != nil {
						log.WithError(err).Warn("----- failed to clone client")
						return
					}

					clt.Stdout = io.Discard
					clt.Stderr = io.Discard
					clt.Stdin = reader
					clt.HostLogin = tc.Host
					clt.Host = node.GetHostname()

					clt.OnShellCreated = func(s *tracessh.Session, c *tracessh.Client, terminal io.ReadWriteCloser) (bool, error) {
						sessionCounter.Inc()

						// we cannot block here otherwise the session will never be
						// fully established. Launch another task so that we can properly
						// clean up the session once it is completed or killed
						g.Go(func() error {
							defer sessionCounter.Dec()

							// hold the session open until either of these
							// contexts is canceled
							select {
							case <-ctx.Done():
							case <-gctx.Done():
							}

							log.Debug("------ exiting")
							// send the exit command to kill the ssh session
							_, err := terminal.Write([]byte("\r\nexit\r\n"))
							return trace.Wrap(err)
						})

						select {
						case <-ctx.Done():
							return true, nil
						case <-gctx.Done():
							return true, nil
						default:
						}

						if _, err := terminal.Write([]byte(cmd + "\r\n")); err != nil {
							return true, trace.Wrap(err)
						}

						return false, nil
					}

					// establish the session and run until killed
					g.Go(func() error {
						return clt.SSH(gctx, nil, false)
					})

				}

				if err := g.Wait(); err != nil && !errors.Is(err, context.Canceled) && !errors.Is(err, io.EOF) {
					log.WithError(err).Warn(" ----- session ending")
				}
			}

			for count := 1; ; count++ {
				logger := log.WithField("count", count)
				select {
				case <-ctx.Done():
					return
				default:
					f()
					if ctx.Err() == nil {
						logger.Warn("---- Session closed, opening a new one")
					}
				}
			}
		}()
	}

	sessions := int64(len(nodes))
	if c.MaxSessions > 0 {
		sessions = int64(c.MaxSessions)
	}

	fmt.Printf("\nAttempting to establish %d sessions\n", sessions)

	for i := int64(0); i < sessions; {
		select {
		case <-spawnTicker.C:
			idx := i % int64(len(nodes))
			worker(i, nodes[idx])
			i++
		case <-statusTicker.C:
			fmt.Printf("\n%d established sessions\n", sessionCounter.Load())
		case <-ctx.Done():
			logrus.Info("--------------------- ctx cancelled - aborting spawning more sessions")
			sessions = -1
		}
	}

	for {
		select {
		case <-ctx.Done():
			fmt.Printf("\nTerminating sesions...\n")
			g.Wait()
			return nil
		case <-statusTicker.C:
			fmt.Printf("\n%d established sessions\n", sessionCounter.Load())
		}
	}
}

type benchMeasure struct {
	ResponseStart time.Time
	End           time.Time
	Error         error
	client        *client.TeleportClient
	command       []string
	interactive   bool
	web           bool
	//pclt          *web.ProxyClient
}

func (m *benchMeasure) work(ctx context.Context, send chan<- *benchMeasure) {

	// do not use parent context that will cancel in flight requests
	// because we give test some time to gracefully wrap up
	// the in-flight connections to avoid extra errors

	if m.web {
		m.Error = m.executeWeb(ctx)
	} else {
		m.Error = m.execute(ctx)
	}

	m.End = time.Now()
	select {
	case send <- m:
	case <-ctx.Done():
		return
	}
}

func (m *benchMeasure) executeWeb(ctx context.Context) error {
	//if m.pclt == nil {
	//	return trace.BadParameter("missing proxy client")
	//}

	//return trace.Wrap(m.pclt.SSH(ctx, m.client, m.command, m.interactive))
	return nil
}

func (m *benchMeasure) execute(ctx context.Context) error {
	if !m.interactive {
		return trace.Wrap(m.client.SSH(ctx, m.command, false))
	}
	config := m.client.Config
	clt, err := client.NewClient(&config)
	if err != nil {
		return trace.Wrap(err)
	}
	reader, writer := io.Pipe()
	defer reader.Close()
	defer writer.Close()
	clt.Stdin = reader
	out := utils.NewSyncBuffer()
	clt.Stdout = out
	clt.Stderr = out
	if err := m.client.SSH(ctx, nil, false); err != nil {
		return trace.Wrap(err)
	}

	if _, err := writer.Write([]byte(strings.Join(m.command, " ") + "\r\nexit\r\n")); err != nil {
		return trace.Wrap(err, "failed to write input")
	}
	return nil
}

// makeTeleportClient creates an instance of a teleport client
func makeTeleportClient(host, login, proxy string) (*client.TeleportClient, error) {
	c := client.Config{
		Host:   host,
		Tracer: tracing.NoopProvider().Tracer("test"),
	}
	path := profile.FullProfilePath("")
	if login != "" {
		c.HostLogin = login
		c.Username = login
	}
	if proxy != "" {
		c.SSHProxyAddr = proxy
	}
	if err := c.LoadProfile(path, proxy); err != nil {
		return nil, trace.Wrap(err)
	}
	tc, err := client.NewClient(&c)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	return tc, nil
}
