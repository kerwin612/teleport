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

package gcp

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"cloud.google.com/go/compute/metadata"
	"github.com/gravitational/trace"

	"github.com/gravitational/teleport/lib/defaults"
)

// contextRoundTripper is a http.RoundTripper that adds a context.Context to
// requests.
type contextRoundTripper struct {
	ctx       context.Context
	transport http.RoundTripper
}

func (rt contextRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := rt.transport.RoundTrip(req.WithContext(rt.ctx))
	return resp, trace.Wrap(err)
}

// getMetadataClient gets an instance metadata client that will use the
// provided context.
func getMetadataClient(ctx context.Context) (*metadata.Client, error) {
	transport, err := defaults.Transport()
	if err != nil {
		return nil, trace.Wrap(err)
	}
	return metadata.NewClient(&http.Client{
		Transport: contextRoundTripper{
			ctx:       ctx,
			transport: transport,
		},
	}), nil
}

func convertMetadataError(err error) error {
	var metadataErr *metadata.Error
	if errors.As(err, &metadataErr) {
		return trace.ReadError(metadataErr.Code, []byte(metadataErr.Message))
	}
	return err
}

// get gets GCP instance metadata from an arbitrary path.
func getMetadata(ctx context.Context, suffix string) (string, error) {
	client, err := getMetadataClient(ctx)
	if err != nil {
		return "", trace.Wrap(err)
	}
	resp, err := client.Get(suffix)
	return resp, trace.Wrap(convertMetadataError(err))
}

// GetIDToken gets an ID token from GCP instance metadata.
func GetIDToken(ctx context.Context) (string, error) {
	audience := "teleport.cluster.local"
	resp, err := getMetadata(ctx, fmt.Sprintf(
		"instance/service-accounts/default/identity?audience=%s&format=full&licenses=FALSE",
		url.QueryEscape(audience),
	))
	return resp, trace.Wrap(err)
}
