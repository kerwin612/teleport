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

import React from 'react';
import { Box, Text, Flex, ButtonBorder, ButtonSecondary } from 'design';
import { DialogContent, DialogHeader } from 'design/Dialog';
import * as Icons from 'design/Icon';

export type HeadlessPromptProps = {
  clientIP: string;
  onApprove(): void;
  onReject(): void;
  onCancel(): void;
};

export function HeadlessPrompt({
  clientIP,
  onApprove,
  onReject,
  onCancel,
}: HeadlessPromptProps) {
  return (
    <Box p={4}>
      <DialogHeader>
        <Text typography="h4">Headless Login Detected</Text>
      </DialogHeader>
      <DialogContent mb={2}>
        <Text>
          Someone has initiated a command from {clientIP}. If it was not you,
          click Reject and contact your administrator.
        </Text>
        <Box mt="5">
          <Flex gap={2}>
            <ButtonBorder
              onClick={e => {
                e.preventDefault();
                onApprove();
              }}
            >
              <Icons.Check fontSize="16px" mr={2} />
              Approve
            </ButtonBorder>
            <ButtonBorder
              onClick={e => {
                e.preventDefault();
                onReject();
              }}
            >
              <Icons.Cross fontSize="16px" mr={2} />
              Reject
            </ButtonBorder>
            <ButtonSecondary
              type="button"
              width={80}
              size="small"
              onClick={onCancel}
            >
              Cancel
            </ButtonSecondary>
          </Flex>
        </Box>
      </DialogContent>
    </Box>
  );
}
