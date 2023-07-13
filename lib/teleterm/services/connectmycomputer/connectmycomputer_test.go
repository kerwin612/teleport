// Copyright 2023 Gravitational, Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package connectmycomputer

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/gravitational/trace"
	"github.com/stretchr/testify/require"

	"github.com/gravitational/teleport/api/types"
	"github.com/gravitational/teleport/lib/client"
	"github.com/gravitational/teleport/lib/teleterm/api/uri"
	"github.com/gravitational/teleport/lib/teleterm/clusters"
)

func TestRoleSetupRun_WithNonLocalUser(t *testing.T) {
	roleSetup, err := NewRoleSetup(&RoleSetupConfig{})
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	oidcUser, err := types.NewUser("alice")
	require.NoError(t, err)
	oidcUser.SetCreatedBy(types.CreatedBy{
		Connector: &types.ConnectorRef{Type: "oidc", ID: "google"},
	})
	accessAndIdentity := &mockAccessAndIdentity{user: oidcUser}
	certManager := &mockCertManager{}

	_, err = roleSetup.Run(ctx, accessAndIdentity, certManager, &clusters.Cluster{URI: uri.NewClusterURI("foo")})
	require.True(t, trace.IsBadParameter(err))
}

func TestRoleSetupRun_UpdateResourcesOnlyOnce(t *testing.T) {
	roleSetup, err := NewRoleSetup(&RoleSetupConfig{})
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	user, err := types.NewUser("alice")
	require.NoError(t, err)
	accessAndIdentity := &mockAccessAndIdentity{user: user, callCounts: make(map[string]int)}
	certManager := &mockCertManager{}

	_, err = roleSetup.Run(ctx, accessAndIdentity, certManager, &clusters.Cluster{URI: uri.NewClusterURI("foo")})
	require.NoError(t, err)

	_, err = roleSetup.Run(ctx, accessAndIdentity, certManager, &clusters.Cluster{URI: uri.NewClusterURI("foo")})
	require.NoError(t, err)

	require.Equal(t, 1, accessAndIdentity.callCounts["UpsertRole"], "expected two runs to update the role only once")
	require.Equal(t, 1, accessAndIdentity.callCounts["UpdateUser"], "expected two runs to update the user only once")
}

type mockAccessAndIdentity struct {
	user       types.User
	role       types.Role
	callCounts map[string]int
	events &fakeEvents{}
}

func (m *mockAccessAndIdentity) GetUser(name string, withSecrets bool) (types.User, error) {
	return m.user, nil
}

func (m *mockAccessAndIdentity) GetRole(ctx context.Context, name string) (types.Role, error) {
	if m.role != nil {
		return m.role, nil
	}
	return nil, trace.NotFound("role not found")
}

func (m *mockAccessAndIdentity) UpsertRole(ctx context.Context, role types.Role) error {
	m.callCounts["UpsertRole"]++
	m.role = role
	return nil
}

// TODO: Watcher in tests?
func (m *mockAccessAndIdentity) NewWatcher(ctx context.Context, watch types.gatch) (types.Watcher, error) {
	return nil, trace.NotImplemented("not implemented")
}

func (m *mockAccessAndIdentity) UpdateUser(ctx context.Context, user types.User) error {
	m.callCounts["UpdateUser"]++
	m.user = user
	return nil
}

type mockCertManager struct{}

func (m *mockCertManager) ReissueUserCerts(context.Context, client.CertCachePolicy, client.ReissueParams) error {
	return nil
}
