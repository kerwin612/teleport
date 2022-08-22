//go:build !bpf || 386
// +build !bpf 386

/*
Copyright 2019 Gravitational, Inc.

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

package bpf

import (
	"github.com/gravitational/teleport/lib/srv"
)

type NOP struct{}

// OpenBPFSession will start monitoring all events within a session and
// emitting them to the Audit Log.
func (n NOP) OpenBPFSession(ctx *srv.ServerContext) (uint64, error) {
	return 0, nil
}

// CloseBPFSession will stop monitoring events for a particular session.
func (n NOP) CloseBPFSession(ctx *srv.ServerContext) error {
	return nil
}

func (n NOP) Close() error {
	return nil
}

// New returns a new NOP service. Note this function does nothing.
func New(config *Config) (NOP, error) {
	return NOP{}, nil
}

// SystemHasBPF returns true if the binary was build with support for BPF
// compiled in.
func SystemHasBPF() bool {
	return false
}
