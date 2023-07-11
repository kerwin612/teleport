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

import React from 'react';

import { SVGIcon } from './SVGIcon';

import type { SVGIconProps } from './common';

export function JenkinsIcon({ size = 64, fill }: SVGIconProps) {
  return (
    <SVGIcon viewBox="0 0 32 32" size={size} fill={fill}>
      <path
        d="M27.295 15.42c0 6.34-5.025 11.48-11.223 11.48S4.85 21.76 4.85 15.42 9.874 3.94 16.072 3.94s11.223 5.14 11.223 11.48"
        fillRule="evenodd"
        fill="#d33833"
      />
      <path
        d="M5.306 18.263S4.494 6.292 15.524 5.95l-.77-1.283-5.986 2-1.7 1.967-1.496 2.865-.855 3.335.257 2.223"
        fill="#ef3d3a"
      />
      <path
        d="M8.385 7.59c-1.97 2.015-3.188 4.797-3.188 7.874s1.22 5.86 3.188 7.873 4.685 3.257 7.687 3.257 5.717-1.243 7.687-3.257 3.188-4.797 3.188-7.873-1.22-5.86-3.188-7.874-4.685-3.257-7.687-3.257-5.717 1.243-7.687 3.257zm-.498 16.235c-2.093-2.14-3.386-5.098-3.386-8.36s1.294-6.22 3.386-8.36 4.988-3.468 8.185-3.467 6.093 1.327 8.185 3.467 3.387 5.098 3.387 8.36-1.294 6.22-3.387 8.36-4.988 3.467-8.185 3.467-6.093-1.327-8.185-3.467"
        fill="#231f20"
      />
      <g fillRule="evenodd">
        <path
          d="m20.796 15.484-1.7.257-2.31.257-1.496.043-1.454-.043-1.112-.342-.983-1.07-.77-2.18-.17-.47-1.026-.342-.6-.983L8.74 9.2l.47-1.24 1.112-.385.898.428.428.94.513-.086.17-.214-.17-.983-.043-1.24.257-1.7-.01-.977.78-1.246 1.368-.983L16.905.477l2.65.385 2.31 1.667 1.07 1.7.684 1.24.17 3.078-.513 2.65-.94 2.35-.898 1.24"
          fill="#f0d6b7"
        />
        <path
          d="m19.342 22.88-6.114.257v1.026l.513 3.59-.257.3L9.21 26.6l-.3-.513-.428-4.83-.983-2.907-.214-.684 3.42-2.35 1.07-.428.94 1.154.812.727.94.3.427.128.513 2.223.385.47.983-.342-.684 1.325 3.72 1.753-.47.257"
          fill="#335061"
        />
        <path
          d="m9.21 7.96 1.112-.385.898.428.428.94.513-.086.128-.513-.257-.983.257-2.352-.214-1.283.77-.898 1.667-1.325-.47-.64-2.352 1.154-.983.77-.556 1.197-.855 1.154-.257 1.368.17 1.454"
          fill="#6d6b6d"
        />
        <path
          d="M10.962 3.94s.64-1.582 3.207-2.35.128-.556.128-.556l-2.78 1.07-1.07 1.07-.47.855.983-.086M9.68 7.66s-.898-2.993 2.523-3.42l-.128-.513-2.35.556-.684 2.223.17 1.454.47-.3"
          fill="#dcd9d8"
        />
        <path
          d="m11.048 11.636.56-.542a.36.36 0 0 0 .295.329c.043.3.17 2.993 2 4.447.168.133-1.368-.214-1.368-.214l-1.368-2.138m7.738-2.695s.1-1.296.45-1.196.35.45.35.45-.848.548-.798.748"
          fill="#f7e4cd"
        />
      </g>
      <path
        d="M22.464 6.077s-.705.15-.77.77.77.128.898.086M17.3 6.12s-.94.128-.94.727 1.07.556 1.368.3"
        fill="#f7e4cd"
      />
      <g fillRule="evenodd">
        <path
          d="M11.475 8.9S9.85 7.916 9.68 8.857s-.556 1.625.257 2.608l-.556-.17-.513-1.325-.17-1.283.983-1.026 1.112.086.64.513.043.64m.77-2.694s.727-3.762 4.404-4.5c3.027-.598 4.617.128 5.216.812 0 0-2.694-3.206-5.26-2.223s-4.446 2.8-4.406 3.943l.043 1.967m9.92-3.25s-1.24-.043-1.283 1.07a.86.86 0 0 0 .085.342s.984-1.112 1.582-.513m-5.813.628s-.213-1.706-1.668-.714c-.94.64-.855 1.54-.684 1.7s.125.515.255.28.087-1.006.558-1.22 1.24-.453 1.54-.055"
          fill="#f7e4cd"
        />
        <path
          d="m12.715 16.04-4.02 1.796s1.667 6.627.812 8.68l-.6-.214-.043-2.523L7.756 19l-.47-1.325 4.2-2.822 1.24 1.197m.404 3.655.57.696v2.565h-.684l-.085-2c0-.214.085-.983.085-.983m.127 3.367-1.924.085.556.385 1.368.214"
          fill="#49728b"
        />
        <g fill="#335061">
          <path d="m19.727 22.923 1.582-.043.385 3.934-1.625.214-.342-4.104" />
          <path d="m20.155 22.923 2.394-.128.983-2.608c0-.128.855-3.59.855-3.59l-1.924-2.01-.385-.342-1.026 1.026v3.976l-.898 3.677" />
        </g>
        <path
          d="m21.224 22.624-1.496.3.214 1.197c.556.257 1.496-.428 1.496-.428m-.172-8.55 2.993 2.223.086-1.026-2.266-2.095-.812.898"
          fill="#49728b"
        />
        <path
          d="m14.627 31.346-.885-3.592-.44-2.65-.073-1.967 4.006-.213h2.493l-.227 4.5.385 3.463-.043.64-3.25.257-1.967-.428"
          fill="#fff"
        />
        <path
          d="M19.17 22.88s-.214 4.447.428 7.6c0 0-1.283.812-3.164 1.026l3.59-.128.428-.257-.513-7.012-.128-1.497"
          fill="#dcd9d8"
        />
        <path
          d="m21.767 26.472 1.667-.47 3.164-.17.47-1.454-.855-2.523-.983-.128-1.368.428-1.313.64-.697-.128-.543.213"
          fill="#fff"
        />
        <path
          d="M21.737 25.616s1.112-.513 1.283-.47l-.47-2.352.556-.214.385 2.48 2.608.128s.513-.983.385-2l.47 1.368.043.77-.684 1.026-.77.17-1.283-.043-.428-.556-1.496.214-.47.17"
          fill="#dcd9d8"
        />
      </g>
      <path
        d="m20.056 22.58-.94-2.394-.983-1.41s.214-.6.513-.6h.983l.94.342-.085 1.582-.428 2.48"
        fill="#fff"
      />
      <path
        d="M20.24 21.77s-1.197-2.31-1.197-2.65c0 0 .214-.513.513-.385s.94.47.94.47v-.812l-1.454-.3-.983.128 1.667 3.934.342.043"
        fill="#dcd9d8"
        fillRule="evenodd"
      />
      <path
        d="m15.01 16.125-1.184-.13-1.112-.342v.385l.543.6 1.7.77"
        fill="#fff"
      />
      <g fillRule="evenodd">
        <path
          d="M13.1 16.254s1.325.556 1.753.428l.043.513-1.197-.256-.727-.513.128-.17"
          fill="#dcd9d8"
        />
        <path
          d="M21.762 18.323c-.725-.02-1.38-.107-1.955-.27.04-.235-.034-.466.025-.636.16-.115.428-.113.67-.14a1.167 1.167 0 0 0-.744-.084c-.006-.163-.08-.264-.123-.392.408-.146 1.37-1.1 1.912-.784.258.15.368 1.01.388 1.426.017.346-.03.696-.173.88"
          fill="#d33833"
        />
      </g>
      <path
        d="M21.762 18.323c-.725-.02-1.38-.107-1.955-.27.04-.235-.034-.466.025-.636.16-.115.428-.113.67-.14a1.167 1.167 0 0 0-.744-.084c-.006-.163-.08-.264-.123-.392.408-.146 1.37-1.1 1.912-.784.258.15.368 1.01.388 1.426.017.346-.03.696-.173.88z"
        fill="none"
        strokeWidth={0.257}
        stroke="#d33833"
      />
      <path
        d="m18.293 17.137-.006.166c-.227.15-.592.147-.84.272.366.016.655.104.905.23l-.016.415c-.415.284-.794.708-1.283.974-.23.126-1.042.45-1.288.393-.14-.032-.152-.205-.207-.368-.118-.35-.39-.903-.415-1.428-.03-.663-.097-1.773.617-1.637a6.38 6.38 0 0 1 1.692.619c.273.15.43.333.844.365"
        fill="#d33833"
        fillRule="evenodd"
      />
      <path
        d="m18.293 17.137-.006.166c-.227.15-.592.147-.84.272.366.016.655.104.905.23l-.016.415c-.415.284-.794.708-1.283.974-.23.126-1.042.45-1.288.393-.14-.032-.152-.205-.207-.368-.118-.35-.39-.903-.415-1.428-.03-.663-.097-1.773.617-1.637a6.38 6.38 0 0 1 1.692.619c.273.15.43.333.844.365z"
        fill="none"
        strokeWidth={0.257}
        stroke="#d33833"
      />
      <path
        d="M18.705 17.928c-.063-.36-.137-.464-.108-.78.963-.642 1.143 1.102.108.78"
        fill="#d33833"
        fillRule="evenodd"
      />
      <path
        d="M18.705 17.928c-.063-.36-.137-.464-.108-.78.963-.642 1.143 1.102.108.78z"
        fill="none"
        strokeWidth={0.257}
        stroke="#d33833"
      />
      <path
        d="M20.1 18.22s-.3-.428-.086-.556.428 0 .556-.214 0-.342.043-.6.257-.3.47-.342.812-.128.898.086l-.257-.77-.513-.17-1.625.94-.086.47v.94m-3.9 1.626-.166-2c-.1-.995.24-.822 1.102-.822.132 0 .8.157.86.257.233.476-.4.37.27.73.556.303 1.538-.184 1.313-.858-.126-.15-.655-.047-.845-.145l-1.002-.52c-.425-.22-1.407-.542-1.86-.234-1.148.78.072 2.732.482 3.546"
        fill="#ef3d3a"
        fillRule="evenodd"
      />
      <path
        d="M16.735 4.483c-1.165-.27-1.744.488-2.098 1.275-.315-.076-.2-.505-.1-.724.2-.574 1.05-1.337 1.736-1.234.296.045.696.315.472.683M22.413 5.8l.055.002c.263.547.5 1.127.823 1.6-.223.518-1.685.977-1.663.046.316-.138.862-.028 1.143-.205-.162-.445-.396-.824-.36-1.453m-5.08.034c.25.458.33.94.686 1.285.16.156.47.346.317.78a1 1 0 0 1-.45.375c-.555.164-1.85.034-1.4-.658.46.02 1.076.298 1.42-.035-.264-.42-.733-1.255-.56-1.746m4.87 4.655c-.836.537-1.768 1.12-3.138.986-.293-.254-.404-.82-.12-1.195.148.254.055.723.468.793.777.133 1.682-.475 2.24-.688.347-.585-.03-.8-.342-1.176-.64-.77-1.497-1.726-1.466-2.88.258-.187.28.286.318.372.334.78 1.174 1.78 1.787 2.45.15.165.4.323.426.432.08.317-.207.696-.174.907m-11.03-.567c-.262-.15-.324-.808-.632-.827-.44-.027-.36.855-.358 1.37-.303-.275-.356-1.12-.134-1.555-.253-.124-.367.137-.507.23.18-1.312 1.92-.6 1.63.783m11.555 1.1c-.4.74-.94 1.556-2.082 1.58-.023-.24-.04-.603.001-.747.873-.084 1.412-.528 2.08-.833m-5.47.48c.728.383 2.067.424 3.057.395.053.217.052.485.054.75-1.273.063-2.777-.25-3.1-1.145m-.14.714c.504 1.265 2.235 1.12 3.695 1.084-.064.164-.204.358-.377.428-.468.2-1.758.335-2.407-.01-.412-.22-.676-.714-.902-1.004-.1-.14-.652-.498-.008-.5"
        fill="#231f20"
        fillRule="evenodd"
      />
      <path
        d="m22.144 19.196-1.858 2.945c.294-.864.42-2.31.464-3.414.615-.288 1.142.065 1.394.47"
        fill="#81b0c4"
        fillRule="evenodd"
      />
      <path
        d="M25.324 22.834c-.662.132-1.127.776-1.772.734.355-.5.976-.7 1.772-.734m.292 1.036c-.54.057-1.173.144-1.72.1.26-.396 1.257-.26 1.72-.1m.187.892c-.606.013-1.36.001-1.936-.047.34-.366 1.543-.136 1.936.047"
        fill="#231f20"
        fillRule="evenodd"
      />
      <path
        d="M21.003 27.206c.087.76.39 1.533.35 2.366-.335.113-.528.212-.977.21l-.098-2.467c.22.015.547-.158.724-.1"
        fill="#dcd9d8"
        fillRule="evenodd"
      />
      <path
        d="M20.026 15.986c-.304.2-.564.447-.856.66-.648.032-1.002-.045-1.478-.417.008-.03.056-.017.057-.053.694.31 1.576-.126 2.277-.19"
        fill="#f0d6b7"
        fillRule="evenodd"
      />
      <path
        d="M16.383 20.716c.19-.826.938-1.254 1.616-1.71.7.89 1.126 2.03 1.595 3.134-1.108-.334-2.24-.876-3.21-1.425"
        fill="#81b0c4"
        fillRule="evenodd"
      />
      <path
        d="M20.28 27.316c-.028.675.066 1.76.098 2.467.45.001.642-.098.977-.21.038-.834-.264-1.605-.35-2.366-.177-.047-.503.125-.724.1zm-6.992-3.84c.296 2.722.725 5 1.512 7.42 1.746.53 3.85.576 5.394.098-.283-1.36-.16-3.017-.325-4.47l-.232-3.303c-1.87-.39-4.513-.09-6.348.252zm6.8-.235c-.016 1.17.052 2.322.142 3.493l1.17-.204c-.135-1.127-.12-2.395-.395-3.392-.32.003-.6-.004-.918.103zm2.277-.188c-.213-.05-.46-.002-.665.002l.4 3.005c.32.01.49-.14.755-.192.014-.878-.077-2.088-.5-2.815zm3.448 3.15c.668-.162 1.087-.98.9-1.82-.125-.564-.348-1.627-.587-1.988-.176-.267-.655-.617-1.037-.372-.622.398-1.717.514-2.17.995.227.757.298 1.796.392 2.755.777.048 1.732-.214 2.378.064-.45.146-1.036.147-1.425.36.318.154 1.064.123 1.55.005zm-6.208-4.06c-.47-1.103-.895-2.246-1.595-3.134-.678.455-1.425.883-1.616 1.71.97.55 2.103 1.09 3.21 1.425zm1.156-3.415c-.044 1.103-.17 2.55-.464 3.414.7-.893 1.267-1.933 1.858-2.945-.252-.404-.78-.757-1.394-.47zm-1.3-.466c-.266-.03-.49.305-.837.16l-.232.268c.763.92 1.1 2.224 1.7 3.304.316-1.038.28-2.175.35-3.308-.434.028-.675-.393-.98-.425zm-.843-1.112c-.028.315.045.418.108.78 1.035.324.854-1.42-.108-.78zm-1.148-.377a6.38 6.38 0 0 0-1.692-.619c-.714-.137-.647.974-.617 1.637.024.525.297 1.08.415 1.428.056.163.068.335.207.368.246.057 1.057-.267 1.288-.393.49-.267.868-.7 1.283-.974l.016-.415a2.208 2.208 0 0 0-.905-.23c.25-.125.614-.123.84-.272l.006-.166c-.414-.032-.57-.216-.844-.365zm-4.26-.774c-.37.376 1.038.888 1.487.916-.003-.238.136-.462.108-.633-.533-.094-1.233-.032-1.594-.283zm4.56.176c-.002.037-.05.023-.057.053.476.372.83.45 1.478.417l.856-.66c-.7.063-1.583.498-2.277.19zm4.186 1.27c-.02-.418-.13-1.276-.388-1.426-.542-.316-1.505.638-1.912.784.044.128.118.23.123.392.24-.06.535-.02.744.084-.242.027-.5.025-.67.14-.06.17.014.4-.025.636.574.162 1.23.248 1.955.27.14-.183.19-.533.173-.88zm-9.37-1.082c-.116-.083-.903-1.106-1-1.064-1.423.56-2.753 1.53-3.942 2.45 1.134 2.432 1.59 5.413 1.672 8.285 1.3.607 2.44 1.483 4.2 1.574l-.506-4.088c-.443-.187-1.078.008-1.492-.058-.004-.5.633-.22.686-.554.04-.254-.35-.273-.223-.673.324.118.494.378.84.475.316-.69-.004-1.912.04-2.5.01-.108.054-.6.297-.514.215.076-.012 1.31.01 1.855.02.503-.06 1 .143 1.306a48.467 48.467 0 0 1 5.275-.432c-.405-.174-.887-.34-1.415-.636-.286-.16-1.188-.497-1.27-.77-.132-.433.346-.664.427-1.035-.86.47-1.027-.45-1.23-1.1-.184-.59-.29-1.03-.334-1.37-.74-.353-1.532-.7-2.17-1.163zm8.617-.94c1.186-.575 1.4 2.15.935 3.026.072.262.32.362.42.597l-2.07 3.463c.502-.313 1.22-.056 1.8-.3.216-.085.372-.58.536-.975.45-1.088.922-2.46 1.132-3.5.047-.237.177-.752.148-.963-.052-.377-.563-.657-.824-.9L22 14.659c-.203.3-.638.5-.804.744zM9.855 4.908c-.565.622-.447 1.786-.378 2.615 1.02-.642 2.376.05 2.364 1.143.488-.013.182-.61.094-.993-.288-1.254.486-2.616.035-3.763-.875.066-1.593.424-2.114.998zM13.897 1.3c-1.28.362-2.918 1.292-3.443 2.44.407-.06.69-.264 1.09-.3.152-.01.35.064.525.02.347-.086.64-.865.903-1.155.256-.283.563-.404.773-.662.135-.065.335-.06.342-.263-.06-.063-.12-.1-.2-.1zm6.657.34C19.227.893 16.98.33 15.568 1.033c-1.14.568-2.68 1.508-3.205 2.7.49 1.15-.145 2.203-.186 3.37-.022.62.292 1.163.316 1.84-.168.277-.68.31-1.036.292-.12-.598-.33-1.27-.944-1.338-.87-.095-1.508.626-1.548 1.38-.047.886.68 2.355 1.712 2.253.398-.04.496-.44.93-.435.235.47-.363.617-.424.952-.016.087.05.426.088.585.187.774.605 1.775 1.016 2.364.522.747 1.546.86 2.65.933.197-.424.922-.39 1.395-.278-.566-.224-1.093-.768-1.53-1.25-.5-.552-1.01-1.145-1.035-1.867.947 1.314 1.73 2.462 3.452 3.04 1.303.437 2.826-.2 3.827-.904.416-.292.664-.756.96-1.18 1.105-1.6 1.62-3.86 1.508-6.06-.047-.907-.045-1.81-.35-2.42-.318-.638-1.394-1.21-2.024-.632-.117-.62.524-1.004 1.276-.78-.536-.692-1.1-1.524-1.862-1.954zm2.473 20.467c1.037-.516 2.975-1.388 3.626.002.24.512.522 1.378.646 1.907.176.746-.2 2.314-.957 2.565-.677.22-1.468.208-2.284.044a1.265 1.265 0 0 1-.278-.364c-.583-.023-1.128.03-1.588.27.044.43-.248.5-.52.59-.203.803.405 1.852.26 2.584-.104.522-.746.602-1.217.7-.015.3.02.532.053.777-.108.397-.592.624-1.05.68-1.508.18-3.798.263-5.25-.26-.405-.993-.724-2.2-1.06-3.335-1.415.15-2.56-.61-3.638-1.11-.374-.173-.9-.27-1.03-.566-.135-.288-.08-.84-.113-1.36-.085-1.33-.16-2.615-.5-3.978-.158-.612-.433-1.15-.625-1.74-.177-.546-.487-1.22-.568-1.766-.12-.807.64-.852 1.126-1.202.75-.54 1.34-.84 2.155-1.328.24-.145.968-.5 1.05-.68.164-.334-.282-.805-.4-1.067a3.184 3.184 0 0 1-.314-1.175 2.268 2.268 0 0 1-1.513-.972c-.517-.758-.876-2.16-.428-3.227.035-.084.2-.25.236-.378.05-.254-.096-.592-.105-.862-.047-1.387.235-2.58 1.168-3 .38-1.5 1.735-2.01 3.013-2.762.478-.28 1.004-.46 1.548-.66 1.95-.718 4.944-.583 6.563.642.686.52 1.784 1.615 2.176 2.41 1.037 2.095.963 5.597.238 8.146-.097.342-.24.845-.436 1.256-.138.287-.565.86-.513 1.115.053.262.975.962 1.173 1.153.356.343 1.032.8 1.087 1.232.06.46-.203 1.092-.336 1.537l-1.38 4.18"
        fill="#231f20"
        fillRule="evenodd"
      />
      <path
        d="M16.107 11.83c.056-.075.366-.19.8.02 0 0-.513.086-.47.94l-.214-.043s-.22-.776-.114-.918"
        fill="#f7e4cd"
        fillRule="evenodd"
      />
      <path
        d="M19.856 19.182a.235.235 0 1 1-.47 0 .235.235 0 0 1 .47 0m.234 1.1a.235.235 0 1 1-.47 0 .235.235 0 0 1 .47 0"
        fill="#1d1919"
        fillRule="evenodd"
      />
    </SVGIcon>
  );
}
