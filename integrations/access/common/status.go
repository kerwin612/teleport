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

package common

import (
	"context"
	"net/http"

	"github.com/gravitational/teleport/api/types"
)

// StatusSink defines a destination for PluginStatus
type StatusSink interface {
	Emit(ctx context.Context, s types.PluginStatus) error
}

func StatusFromStatusCode(httpCode int) types.PluginStatus {
	var code types.PluginStatusCode
	switch {
	case httpCode == http.StatusUnauthorized:
		code = types.PluginStatusCode_UNAUTHORIZED
	case httpCode >= 200 && httpCode < 400:
		code = types.PluginStatusCode_RUNNING
	default:
		code = types.PluginStatusCode_OTHER_ERROR
	}
	return &types.PluginStatusV1{Code: code}
}
