/**
 * Copyright 2023 Gravitational, Inc
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import { SendPendingHeadlessAuthenticationRequest } from 'teleterm/services/tshdEvents';
import { MainProcessClient } from 'teleterm/types';
import { ModalsService } from 'teleterm/ui/services/modals';
import type * as types from 'teleterm/services/tshd/types';

export class HeadlessAuthenticationService {
  constructor(
    private mainProcessClient: MainProcessClient,
    private modalsService: ModalsService,
    private tshClient: types.TshClient
  ) {}

  sendPendingHeadlessAuthentication(
    request: SendPendingHeadlessAuthenticationRequest
  ) {
    this.mainProcessClient.forceFocusWindow();
    this.modalsService.openImportantDialog({
      kind: 'headless-authn',
      rootClusterURI: request.rootClusterUri,
      headlessAuthenticationID: request.headlessAuthenticationId,
      headlessAuthenticationClientIP: request.headlessAuthenticationClientIp,
    });
  }

  async updateHeadlessAuthenticationState(
    params: types.UpdateHeadlessAuthenticationStateParams
  ): Promise<void> {
    return this.tshClient.updateHeadlessAuthenticationState(params);
  }
}
