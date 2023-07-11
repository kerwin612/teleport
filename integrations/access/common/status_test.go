/**
 * Copyright 2023 Gravitational, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package common

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/gravitational/teleport/api/types"
	"github.com/stretchr/testify/require"
)

func TestStatusFromStatusCode(t *testing.T) {
	testCases := []struct {
		httpCode int
		want     types.PluginStatusCode
	}{
		{
			httpCode: http.StatusOK,
			want:     types.PluginStatusCode_RUNNING,
		},
		{
			httpCode: http.StatusNoContent,
			want:     types.PluginStatusCode_RUNNING,
		},

		{
			httpCode: http.StatusUnauthorized,
			want:     types.PluginStatusCode_UNAUTHORIZED,
		},
		{
			httpCode: http.StatusInternalServerError,
			want:     types.PluginStatusCode_OTHER_ERROR,
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%d", tc.httpCode), func(t *testing.T) {
			require.Equal(t, tc.want, StatusFromStatusCode(tc.httpCode).GetCode())
		})
	}
}
