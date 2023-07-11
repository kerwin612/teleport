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

package types

import (
	"testing"
	"time"

	"github.com/gravitational/trace"
	"github.com/stretchr/testify/require"
)

func TestPluginWithoutSecrets(t *testing.T) {
	spec := PluginSpecV1{
		Settings: &PluginSpecV1_SlackAccessPlugin{
			SlackAccessPlugin: &PluginSlackAccessSettings{
				FallbackChannel: "#access-requests",
			},
		},
	}

	creds := &PluginCredentialsV1{
		Credentials: &PluginCredentialsV1_Oauth2AccessToken{
			Oauth2AccessToken: &PluginOAuth2AccessTokenCredentials{
				AccessToken:  "access_token",
				RefreshToken: "refresh_token",
				Expires:      time.Now().UTC(),
			},
		},
	}

	plugin := NewPluginV1(Metadata{Name: "foobar"}, spec, creds)
	plugin = plugin.WithoutSecrets().(*PluginV1)
	require.Nil(t, plugin.Credentials)
}

func TestPluginOpenAIValidation(t *testing.T) {
	spec := PluginSpecV1{
		Settings: &PluginSpecV1_Openai{},
	}
	testCases := []struct {
		name      string
		creds     *PluginCredentialsV1
		assertErr require.ErrorAssertionFunc
	}{
		{
			name:  "no credentials",
			creds: nil,
			assertErr: func(t require.TestingT, err error, args ...any) {
				require.Error(t, err)
				require.True(t, trace.IsBadParameter(err))
				require.Contains(t, err.Error(), "credentials must be set")
			},
		},
		{
			name:  "no credentials inner",
			creds: &PluginCredentialsV1{},
			assertErr: func(t require.TestingT, err error, args ...any) {
				require.Error(t, err)
				require.True(t, trace.IsBadParameter(err))
				require.Contains(t, err.Error(), "must be used with the bearer token credential type")
			},
		},
		{
			name: "invalid credential type (oauth2)",
			creds: &PluginCredentialsV1{
				Credentials: &PluginCredentialsV1_Oauth2AccessToken{},
			},
			assertErr: func(t require.TestingT, err error, args ...any) {
				require.Error(t, err)
				require.True(t, trace.IsBadParameter(err))
				require.Contains(t, err.Error(), "must be used with the bearer token credential type")
			},
		},
		{
			name: "valid credentials (token)",
			creds: &PluginCredentialsV1{
				Credentials: &PluginCredentialsV1_BearerToken{
					BearerToken: &PluginBearerTokenCredentials{
						Token: "xxx-abc",
					},
				},
			},
			assertErr: func(t require.TestingT, err error, args ...any) {
				require.NoError(t, err)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			plugin := NewPluginV1(Metadata{Name: "foobar"}, spec, tc.creds)
			tc.assertErr(t, plugin.CheckAndSetDefaults())
		})
	}
}

func TestPluginOpsgenieValidation(t *testing.T) {
	testCases := []struct {
		name      string
		settings  *PluginSpecV1_Opsgenie
		creds     *PluginCredentialsV1
		assertErr require.ErrorAssertionFunc
	}{
		{
			name: "no settings",
			settings: &PluginSpecV1_Opsgenie{
				Opsgenie: nil,
			},
			creds: nil,
			assertErr: func(t require.TestingT, err error, args ...any) {
				require.True(t, trace.IsBadParameter(err))
				require.Contains(t, err.Error(), "missing opsgenie settings")
			},
		},
		{
			name: "no api endpint",
			settings: &PluginSpecV1_Opsgenie{
				Opsgenie: &PluginOpsgenieAccessSettings{},
			},
			creds: nil,
			assertErr: func(t require.TestingT, err error, args ...any) {
				require.True(t, trace.IsBadParameter(err))
				require.Contains(t, err.Error(), "api endpoint url must be set")
			},
		},
		{
			name: "no static credentials",
			settings: &PluginSpecV1_Opsgenie{
				Opsgenie: &PluginOpsgenieAccessSettings{
					ApiEndpoint: "https://test.opsgenie.com",
				},
			},
			assertErr: func(t require.TestingT, err error, args ...any) {
				require.True(t, trace.IsBadParameter(err))
				require.Contains(t, err.Error(), "must be used with the static credentials ref type")
			},
		},
		{
			name: "static credentials labels not defined",
			settings: &PluginSpecV1_Opsgenie{
				Opsgenie: &PluginOpsgenieAccessSettings{
					ApiEndpoint: "https://test.opsgenie.com",
				},
			},
			creds: &PluginCredentialsV1{
				Credentials: &PluginCredentialsV1_StaticCredentialsRef{
					&PluginStaticCredentialsRef{
						Labels: map[string]string{},
					},
				},
			},
			assertErr: func(t require.TestingT, err error, args ...any) {
				require.True(t, trace.IsBadParameter(err))
				require.Contains(t, err.Error(), "labels must be specified")
			},
		},
		{
			name: "valid credentials (static credentials)",
			settings: &PluginSpecV1_Opsgenie{
				Opsgenie: &PluginOpsgenieAccessSettings{
					ApiEndpoint: "https://test.opsgenie.com",
				},
			},
			creds: &PluginCredentialsV1{
				Credentials: &PluginCredentialsV1_StaticCredentialsRef{
					&PluginStaticCredentialsRef{
						Labels: map[string]string{
							"label1": "value1",
						},
					},
				},
			},
			assertErr: func(t require.TestingT, err error, args ...any) {
				require.NoError(t, err)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			plugin := NewPluginV1(Metadata{Name: "foobar"}, PluginSpecV1{
				Settings: tc.settings,
			}, tc.creds)
			tc.assertErr(t, plugin.CheckAndSetDefaults())
		})
	}
}

func TestPluginOktaValidation(t *testing.T) {
	testCases := []struct {
		name      string
		settings  *PluginSpecV1_Okta
		creds     *PluginCredentialsV1
		assertErr require.ErrorAssertionFunc
	}{
		{
			name: "no settings",
			settings: &PluginSpecV1_Okta{
				Okta: nil,
			},
			creds: nil,
			assertErr: func(t require.TestingT, err error, args ...any) {
				require.True(t, trace.IsBadParameter(err))
				require.Contains(t, err.Error(), "missing Okta settings")
			},
		},
		{
			name: "no org URL",
			settings: &PluginSpecV1_Okta{
				Okta: &PluginOktaSettings{},
			},
			creds: nil,
			assertErr: func(t require.TestingT, err error, args ...any) {
				require.True(t, trace.IsBadParameter(err))
				require.Contains(t, err.Error(), "org_url must be set")
			},
		},
		{
			name: "no credentials inner",
			settings: &PluginSpecV1_Okta{
				Okta: &PluginOktaSettings{
					OrgUrl: "https://test.okta.com",
				},
			},
			creds: &PluginCredentialsV1{},
			assertErr: func(t require.TestingT, err error, args ...any) {
				require.True(t, trace.IsBadParameter(err))
				require.Contains(t, err.Error(), "must be used with the static credentials ref type")
			},
		},
		{
			name: "invalid credential type (oauth2)",
			settings: &PluginSpecV1_Okta{
				Okta: &PluginOktaSettings{
					OrgUrl: "https://test.okta.com",
				},
			},
			creds: &PluginCredentialsV1{
				Credentials: &PluginCredentialsV1_Oauth2AccessToken{},
			},
			assertErr: func(t require.TestingT, err error, args ...any) {
				require.True(t, trace.IsBadParameter(err))
				require.Contains(t, err.Error(), "must be used with the static credentials ref type")
			},
		},
		{
			name: "invalid credentials (static credentials)",
			settings: &PluginSpecV1_Okta{
				Okta: &PluginOktaSettings{
					OrgUrl: "https://test.okta.com",
				},
			},
			creds: &PluginCredentialsV1{
				Credentials: &PluginCredentialsV1_StaticCredentialsRef{
					&PluginStaticCredentialsRef{
						Labels: map[string]string{},
					},
				},
			},
			assertErr: func(t require.TestingT, err error, args ...any) {
				require.True(t, trace.IsBadParameter(err))
				require.Contains(t, err.Error(), "labels must be specified")
			},
		},
		{
			name: "valid credentials (static credentials)",
			settings: &PluginSpecV1_Okta{
				Okta: &PluginOktaSettings{
					OrgUrl: "https://test.okta.com",
				},
			},
			creds: &PluginCredentialsV1{
				Credentials: &PluginCredentialsV1_StaticCredentialsRef{
					&PluginStaticCredentialsRef{
						Labels: map[string]string{
							"label1": "value1",
						},
					},
				},
			},
			assertErr: func(t require.TestingT, err error, args ...any) {
				require.NoError(t, err)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			plugin := NewPluginV1(Metadata{Name: "foobar"}, PluginSpecV1{
				Settings: tc.settings,
			}, tc.creds)
			tc.assertErr(t, plugin.CheckAndSetDefaults())
		})
	}
}

func TestPluginJamfValidation(t *testing.T) {
	testCases := []struct {
		name      string
		settings  *PluginSpecV1_Jamf
		creds     *PluginCredentialsV1
		assertErr require.ErrorAssertionFunc
	}{
		{
			name: "no settings",
			settings: &PluginSpecV1_Jamf{
				Jamf: nil,
			},
			creds: nil,
			assertErr: func(t require.TestingT, err error, args ...any) {
				require.True(t, trace.IsBadParameter(err))
				require.Contains(t, err.Error(), "missing Jamf settings")
			},
		},
		{
			name: "no api Endpoint",
			settings: &PluginSpecV1_Jamf{
				Jamf: &PluginJamfSettings{
					JamfSpec: &JamfSpecV1{},
				},
			},
			creds: nil,
			assertErr: func(t require.TestingT, err error, args ...any) {
				require.True(t, trace.IsBadParameter(err))
				require.Contains(t, err.Error(), "api endpoint must be set")
			},
		},
		{
			name: "no credentials inner",
			settings: &PluginSpecV1_Jamf{
				Jamf: &PluginJamfSettings{
					JamfSpec: &JamfSpecV1{
						ApiEndpoint: "https://api.testjamfserver.com",
					},
				},
			},
			creds: &PluginCredentialsV1{},
			assertErr: func(t require.TestingT, err error, args ...any) {
				require.True(t, trace.IsBadParameter(err))
				require.Contains(t, err.Error(), "must be used with the static credentials ref type")
			},
		},
		{
			name: "invalid credential type (oauth2)",
			settings: &PluginSpecV1_Jamf{
				Jamf: &PluginJamfSettings{
					JamfSpec: &JamfSpecV1{
						ApiEndpoint: "https://api.testjamfserver.com",
					},
				},
			},
			creds: &PluginCredentialsV1{
				Credentials: &PluginCredentialsV1_Oauth2AccessToken{},
			},
			assertErr: func(t require.TestingT, err error, args ...any) {
				require.True(t, trace.IsBadParameter(err))
				require.Contains(t, err.Error(), "must be used with the static credentials ref type")
			},
		},
		{
			name: "invalid credentials (static credentials)",
			settings: &PluginSpecV1_Jamf{
				Jamf: &PluginJamfSettings{
					JamfSpec: &JamfSpecV1{
						ApiEndpoint: "https://api.testjamfserver.com",
					},
				},
			},
			creds: &PluginCredentialsV1{
				Credentials: &PluginCredentialsV1_StaticCredentialsRef{
					&PluginStaticCredentialsRef{
						Labels: map[string]string{},
					},
				},
			},
			assertErr: func(t require.TestingT, err error, args ...any) {
				require.True(t, trace.IsBadParameter(err))
				require.Contains(t, err.Error(), "labels must be specified")
			},
		},
		{
			name: "valid credentials (static credentials)",
			settings: &PluginSpecV1_Jamf{
				Jamf: &PluginJamfSettings{
					JamfSpec: &JamfSpecV1{
						ApiEndpoint: "https://api.testjamfserver.com",
					},
				},
			},
			creds: &PluginCredentialsV1{
				Credentials: &PluginCredentialsV1_StaticCredentialsRef{
					&PluginStaticCredentialsRef{
						Labels: map[string]string{
							"label1": "value1",
						},
					},
				},
			},
			assertErr: func(t require.TestingT, err error, args ...any) {
				require.NoError(t, err)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			plugin := NewPluginV1(Metadata{Name: "foobar"}, PluginSpecV1{
				Settings: tc.settings,
			}, tc.creds)
			tc.assertErr(t, plugin.CheckAndSetDefaults())
		})
	}
}

func TestPluginMattermostValidation(t *testing.T) {
	defaultSettings := &PluginSpecV1_Mattermost{
		Mattermost: &PluginMattermostSettings{
			ServerUrl: "https://test.mattermost.com",
			Team:      "team-llama",
			Channel:   "teleport",
		},
	}

	testCases := []struct {
		name      string
		settings  *PluginSpecV1_Mattermost
		creds     *PluginCredentialsV1
		assertErr require.ErrorAssertionFunc
	}{
		{
			name: "no settings",
			settings: &PluginSpecV1_Mattermost{
				Mattermost: nil,
			},
			creds: nil,
			assertErr: func(t require.TestingT, err error, args ...any) {
				require.True(t, trace.IsBadParameter(err))
				require.Contains(t, err.Error(), "missing Mattermost settings")
			},
		},
		{
			name: "no server url",
			settings: &PluginSpecV1_Mattermost{
				Mattermost: &PluginMattermostSettings{},
			},
			creds: nil,
			assertErr: func(t require.TestingT, err error, args ...any) {
				require.True(t, trace.IsBadParameter(err))
				require.Contains(t, err.Error(), "server url is required")
			},
		},
		{
			name: "no team",
			settings: &PluginSpecV1_Mattermost{
				Mattermost: &PluginMattermostSettings{
					ServerUrl: "https://test.mattermost.com",
				},
			},
			creds: nil,
			assertErr: func(t require.TestingT, err error, args ...any) {
				require.True(t, trace.IsBadParameter(err))
				require.Contains(t, err.Error(), "team is required")
			},
		},
		{
			name: "no channel",
			settings: &PluginSpecV1_Mattermost{
				Mattermost: &PluginMattermostSettings{
					ServerUrl: "https://test.mattermost.com",
					Team:      "team-llama",
				},
			},
			creds: nil,
			assertErr: func(t require.TestingT, err error, args ...any) {
				require.True(t, trace.IsBadParameter(err))
				require.Contains(t, err.Error(), "channel is required")
			},
		},
		{
			name:     "no credentials inner",
			settings: defaultSettings,
			creds:    &PluginCredentialsV1{},
			assertErr: func(t require.TestingT, err error, args ...any) {
				require.True(t, trace.IsBadParameter(err))
				require.Contains(t, err.Error(), "must be used with the static credentials ref type")
			},
		},
		{
			name:     "invalid credential type (oauth2)",
			settings: defaultSettings,
			creds: &PluginCredentialsV1{
				Credentials: &PluginCredentialsV1_Oauth2AccessToken{},
			},
			assertErr: func(t require.TestingT, err error, args ...any) {
				require.True(t, trace.IsBadParameter(err))
				require.Contains(t, err.Error(), "must be used with the static credentials ref type")
			},
		},
		{
			name:     "no labels for credentials",
			settings: defaultSettings,
			creds: &PluginCredentialsV1{
				Credentials: &PluginCredentialsV1_StaticCredentialsRef{
					&PluginStaticCredentialsRef{
						Labels: map[string]string{},
					},
				},
			},
			assertErr: func(t require.TestingT, err error, args ...any) {
				require.True(t, trace.IsBadParameter(err))
				require.Contains(t, err.Error(), "labels must be specified")
			},
		},
		{
			name:     "valid settings",
			settings: defaultSettings,
			creds: &PluginCredentialsV1{
				Credentials: &PluginCredentialsV1_StaticCredentialsRef{
					&PluginStaticCredentialsRef{
						Labels: map[string]string{
							"label1": "value1",
						},
					},
				},
			},
			assertErr: func(t require.TestingT, err error, args ...any) {
				require.NoError(t, err)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			plugin := NewPluginV1(Metadata{Name: "foobar"}, PluginSpecV1{
				Settings: tc.settings,
			}, tc.creds)
			tc.assertErr(t, plugin.CheckAndSetDefaults())
		})
	}
}
