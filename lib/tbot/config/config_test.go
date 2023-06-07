/*
Copyright 2022 Gravitational, Inc.

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

package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/gravitational/teleport/api/types"
)

func TestConfigCLIOnlySample(t *testing.T) {
	// Test the sample config generated by `tctl bots add ...`
	cf := CLIConf{
		DestinationDir: "/tmp/foo",
		Token:          "foo",
		CAPins:         []string{"abc123"},
		AuthServer:     "auth.example.com",
		DiagAddr:       "127.0.0.1:1337",
		Debug:          true,
		JoinMethod:     string(types.JoinMethodToken),
	}
	cfg, err := FromCLIConf(&cf)
	require.NoError(t, err)

	require.Equal(t, cf.AuthServer, cfg.AuthServer)

	require.NotNil(t, cfg.Onboarding)

	token, err := cfg.Onboarding.Token()
	require.NoError(t, err)
	require.Equal(t, cf.Token, token)
	require.Equal(t, cf.CAPins, cfg.Onboarding.CAPins)

	// Storage is still default
	storageDest, err := cfg.Storage.GetDestination()
	require.NoError(t, err)
	storageImpl, ok := storageDest.(*DestinationDirectory)
	require.True(t, ok)
	require.Equal(t, defaultStoragePath, storageImpl.Path)

	// A single default destination should exist
	require.Len(t, cfg.Destinations, 1)
	dest := cfg.Destinations[0]

	// We have 3 required/default templates.
	require.Len(t, dest.Configs, 3)
	template := dest.Configs[0]
	require.NotNil(t, template.SSHClient)

	destImpl, err := dest.GetDestination()
	require.NoError(t, err)
	destImplReal, ok := destImpl.(*DestinationDirectory)
	require.True(t, ok)

	require.Equal(t, cf.DestinationDir, destImplReal.Path)
	require.Equal(t, cf.Debug, cfg.Debug)
	require.Equal(t, cf.DiagAddr, cfg.DiagAddr)
}

func TestConfigFile(t *testing.T) {
	configData := fmt.Sprintf(exampleConfigFile, "foo")
	cfg, err := ReadConfig(strings.NewReader(configData))
	require.NoError(t, err)

	require.Equal(t, "auth.example.com", cfg.AuthServer)
	require.Equal(t, time.Minute*5, cfg.RenewalInterval)

	require.NotNil(t, cfg.Onboarding)

	token, err := cfg.Onboarding.Token()
	require.NoError(t, err)
	require.Equal(t, "foo", token)
	require.ElementsMatch(t, []string{"sha256:abc123"}, cfg.Onboarding.CAPins)

	storage, err := cfg.Storage.GetDestination()
	require.NoError(t, err)

	_, ok := storage.(*DestinationMemory)
	require.True(t, ok)

	require.Len(t, cfg.Destinations, 1)
	destination := cfg.Destinations[0]

	require.Len(t, destination.Configs, 1)
	template := destination.Configs[0]
	templateImpl, err := template.GetConfigTemplate()
	require.NoError(t, err)
	sshTemplate, ok := templateImpl.(*TemplateSSHClient)
	require.True(t, ok)
	require.Equal(t, uint16(1234), sshTemplate.ProxyPort)

	destImpl, err := destination.GetDestination()
	require.NoError(t, err)
	destImplReal, ok := destImpl.(*DestinationDirectory)
	require.True(t, ok)
	require.Equal(t, "/tmp/foo", destImplReal.Path)

	require.True(t, cfg.Debug)
	require.Equal(t, "127.0.0.1:1337", cfg.DiagAddr)
}

func TestLoadTokenFromFile(t *testing.T) {
	tokenDir := t.TempDir()
	tokenFile := filepath.Join(tokenDir, "token")
	require.NoError(t, os.WriteFile(tokenFile, []byte("xxxyyy"), 0660))

	configData := fmt.Sprintf(exampleConfigFile, tokenFile)
	cfg, err := ReadConfig(strings.NewReader(configData))
	require.NoError(t, err)

	token, err := cfg.Onboarding.Token()
	require.NoError(t, err)
	require.Equal(t, token, "xxxyyy")
}

const exampleConfigFile = `
auth_server: auth.example.com
renewal_interval: 5m
debug: true
diag_addr: 127.0.0.1:1337
onboarding:
  token: %s
  ca_pins:
    - sha256:abc123
storage:
  memory: {}
destinations:
  - directory:
      path: /tmp/foo
    configs:
      - ssh_client:
          proxy_port: 1234
`

func TestStorageConfigFromCLIConf(t *testing.T) {
	tests := []struct {
		in      string
		want    *StorageConfig
		wantErr bool
	}{
		{
			in: "/absolute/dir",
			want: &StorageConfig{
				DestinationMixin: DestinationMixin{
					Directory: &DestinationDirectory{
						Path: "/absolute/dir",
					},
				},
			},
		},
		{
			in: "relative/dir",
			want: &StorageConfig{
				DestinationMixin: DestinationMixin{
					Directory: &DestinationDirectory{
						Path: "relative/dir",
					},
				},
			},
		},
		{
			in: "./relative/dir",
			want: &StorageConfig{
				DestinationMixin: DestinationMixin{
					Directory: &DestinationDirectory{
						Path: "./relative/dir",
					},
				},
			},
		},
		{
			in: "file:///absolute/dir",
			want: &StorageConfig{
				DestinationMixin: DestinationMixin{
					Directory: &DestinationDirectory{
						Path: "/absolute/dir",
					},
				},
			},
		},
		{
			in: "file:/absolute/dir",
			want: &StorageConfig{
				DestinationMixin: DestinationMixin{
					Directory: &DestinationDirectory{
						Path: "/absolute/dir",
					},
				},
			},
		},
		{
			in:      "file://host/absolute/dir",
			wantErr: true,
		},
		{
			in: "memory://",
			want: &StorageConfig{
				DestinationMixin: DestinationMixin{
					Memory: &DestinationMemory{},
				},
			},
		},
		{
			in:      "memory://foo/bar",
			wantErr: true,
		},
		{
			in:      "foobar://",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got, err := storageConfigFromCLIConf(tt.in)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}
