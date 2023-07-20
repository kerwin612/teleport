/**
 * Copyright 2021 Gravitational, Inc.
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

import React, { useState } from 'react';
import { ButtonIcon, Box, Text, ButtonSecondary } from 'design';
import Dialog from 'design/Dialog';
import { DialogContent, DialogHeader, DialogFooter } from 'design/Dialog';
import { PromptWebauthn } from '../../ClusterConnect/ClusterLogin/FormLogin/PromptWebauthn';
import * as Icons from 'design/Icon';

export type HeadlessPromptProps = {
  clientIP: string;
  onApprove(): void;
  onReject(): void;
  onAbort(): void;
};

export function HeadlessPrompt({
  clientIP,
  onApprove,
  onReject,
  onAbort,
}: HeadlessPromptProps) {
  const [waitForMfa, setWaitForMfa] = useState(false);

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
      <Box p={4}>
        <DialogHeader
          justifyContent="space-between"
          mb={0}
          alignItems="baseline"
        >
          <Text typography="h4" bold>
            Headless Login Detected
          </Text>
          <ButtonIcon
            type="button"
            onClick={onAbort}
            color="text.slightlyMuted"
          >
            <Icons.Close fontSize={5} />
          </ButtonIcon>
        </DialogHeader>
        <DialogContent>
          <Text typography="body1" color="text.slightlyMuted">
            Someone has initiated a command from {clientIP}. If it was not you,
            click Reject and contact your administrator.
          </Text>
        </DialogContent>
        {!waitForMfa && (
          <DialogFooter>
            <ButtonSecondary
              autoFocus
              mr={3}
              type="submit"
              onClick={e => {
                e.preventDefault();
                setWaitForMfa(true);
                onApprove();
              }}
            >
              Approve
            </ButtonSecondary>
            <ButtonSecondary
              type="button"
              onClick={e => {
                e.preventDefault();
                onReject();
              }}
            >
              Reject
            </ButtonSecondary>
          </DialogFooter>
        )}
        {waitForMfa && (
          <DialogContent mb={2}>
            <Text>Complete MFA verification to approve the Headless Login</Text>
            <PromptWebauthn prompt={'tap'} onCancel={onAbort} />
          </DialogContent>
        )}
      </Box>
    </Dialog>
  );
}
