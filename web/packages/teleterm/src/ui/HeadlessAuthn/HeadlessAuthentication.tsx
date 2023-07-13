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

import React from 'react';

import Dialog from 'design/Dialog';

import { useAppContext } from 'teleterm/ui/appContextProvider';
import { RootClusterUri } from 'teleterm/ui/uri';
import { HeadlessAuthenticationState } from 'gen-proto-js/teleport/lib/teleterm/v1/service_pb';

import { HeadlessPrompt } from './HeadlessPrompt';

export function HeadlessAuthentication(props: HeadlessAuthenticationProps) {
  const { clustersService } = useAppContext();

  function handleHeadlessApprove(): void {
    // TODO prompt webauthn while waiting for updateHeadlessAuthenticationState request, which prompts
    // for webauthn from the tshd daemon.
    clustersService.updateHeadlessAuthenticationState({
      rootClusterURI: props.rootClusterURI,
      headlessAuthenticationID: props.headlessAuthenticationID,
      state: HeadlessAuthenticationState.HEADLESS_AUTHENTICATION_STATE_APPROVED,
    });

    return props.onSuccess();
  }

  function handleHeadlessReject(): void {
    clustersService.updateHeadlessAuthenticationState({
      rootClusterURI: props.rootClusterURI,
      headlessAuthenticationID: props.headlessAuthenticationID,
      state: HeadlessAuthenticationState.HEADLESS_AUTHENTICATION_STATE_DENIED,
    });

    return props.onSuccess();
  }

  return (
    <Dialog
      dialogCss={() => ({
        maxWidth: '480px',
        width: '100%',
        padding: '0',
      })}
      disableEscapeKeyDown={false}
      open={true}
    >
      <HeadlessPrompt
        clientIP={props.clientIP}
        onApprove={handleHeadlessApprove}
        onReject={handleHeadlessReject}
        onCancel={props.onCancel}
      />
    </Dialog>
  );
}

interface HeadlessAuthenticationProps {
  rootClusterURI: RootClusterUri;
  headlessAuthenticationID: string;
  clientIP: string;
  onCancel(): void;
  onSuccess(): void;
}
