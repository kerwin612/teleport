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

import type { Desktop, WindowsDesktopService } from './types';

export function makeDesktop(json): Desktop {
  const { os, name, addr, host_id } = json;

  const labels = json.labels || [];
  const logins = json.logins || [];

  return {
    kind: 'windows_desktop',
    os,
    name,
    addr,
    labels,
    host_id,
    logins,
  };
}

export function makeDesktopService(json): WindowsDesktopService {
  const { name, hostname, addr } = json;

  const labels = json.labels || [];

  return {
    hostname,
    addr,
    labels,
    name,
  };
}
