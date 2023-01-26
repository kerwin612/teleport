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

import type { SVGIconProps } from './common';

export function KubernetesIcon({ size = 20, fill = 'white' }: SVGIconProps) {
  return (
    <svg
      viewBox="0 0 20 20"
      xmlns="http://www.w3.org/2000/svg"
      width={size}
      height={size}
      fill={fill}
    >
      <path d="M8.50323 11.9581L8.50886 11.9662L7.67636 13.9769C6.88511 13.465 6.28198 12.7212 5.95698 11.8425L5.94761 11.8125L8.09573 11.4481L8.09886 11.4525C8.11886 11.4487 8.14198 11.4462 8.16511 11.4462C8.36761 11.4462 8.53198 11.6106 8.53198 11.8131C8.53198 11.865 8.52136 11.915 8.50136 11.96L8.50198 11.9575L8.50323 11.9581ZM7.80948 10.1844C7.96636 10.1406 8.07948 9.99875 8.07948 9.83062C8.07948 9.72062 8.03073 9.62187 7.95448 9.55437L7.95386 9.55375L7.95573 9.54437L6.32136 8.0825C5.92323 8.72062 5.68698 9.49562 5.68698 10.3256C5.68698 10.4906 5.69636 10.6537 5.71448 10.8144L5.71261 10.795L7.80761 10.1906L7.80948 10.1831V10.1844ZM8.76323 8.53437C8.82323 8.57812 8.89823 8.605 8.97948 8.605C9.17636 8.605 9.33761 8.44937 9.34573 8.25437V8.25375L9.35386 8.24937L9.47886 6.06625C8.49386 6.18437 7.62761 6.62125 6.97011 7.26875L6.97073 7.26812L8.76011 8.5375L8.76323 8.53562V8.53437ZM9.39698 10.8256L9.99948 11.1162L10.6014 10.8269L10.7514 10.1769L10.3345 9.6575H9.66448L9.24761 10.1769L9.39698 10.8263V10.8256ZM10.647 8.24687C10.6557 8.4425 10.8164 8.59812 11.0132 8.59812C11.0951 8.59812 11.1701 8.57125 11.2314 8.52625L11.2301 8.52687L11.237 8.52938L13.0151 7.26875C12.3601 6.62687 11.5007 6.19187 10.5432 6.06938L10.522 6.06687L10.6451 8.24625L10.647 8.24687ZM19.7101 13.2219L14.8995 19.2062C14.6532 19.5094 14.2801 19.7013 13.8626 19.7013C13.8614 19.7013 13.8601 19.7013 13.8595 19.7013L6.14198 19.7037C6.14198 19.7037 6.14136 19.7037 6.14073 19.7037C5.72261 19.7037 5.34886 19.5112 5.10448 19.2094L5.10261 19.2069L0.289482 13.2238C0.109482 13 0.000732422 12.7131 0.000732422 12.4C0.000732422 12.2962 0.0126074 12.195 0.0357324 12.0981L0.0338574 12.1069L1.75073 4.64438C1.84073 4.2525 2.10511 3.92438 2.47011 3.75L9.42198 0.4275C9.59073 0.345 9.78948 0.296875 9.99886 0.296875C10.2082 0.296875 10.407 0.345 10.5839 0.43125L10.5757 0.4275L17.5301 3.74813C17.8951 3.9225 18.1595 4.25125 18.2495 4.6425L19.9676 12.105C20.0576 12.4969 19.9632 12.9075 19.7107 13.2219H19.7101ZM16.9695 11.5075C16.9345 11.4994 16.8839 11.4856 16.8489 11.4794C16.7039 11.4519 16.5864 11.4587 16.4495 11.4475C16.1576 11.4169 15.9176 11.3919 15.7039 11.3244C15.6164 11.2912 15.5539 11.1869 15.5239 11.1444L15.3564 11.095C15.3882 10.8744 15.407 10.62 15.407 10.3612C15.407 9.9325 15.357 9.515 15.262 9.115L15.2695 9.15125C15.1095 8.46688 14.8401 7.86437 14.4776 7.33L14.4895 7.34875C14.5326 7.30937 14.6145 7.23812 14.637 7.21625C14.6439 7.14125 14.6376 7.06375 14.7151 6.98125C14.8795 6.82687 15.0851 6.69937 15.3345 6.54625C15.4526 6.47625 15.562 6.43187 15.6801 6.34437C15.707 6.32437 15.7432 6.2925 15.772 6.27C15.972 6.11062 16.0176 5.83687 15.8745 5.65687C15.7314 5.47687 15.4526 5.46 15.2539 5.61937C15.2257 5.64187 15.187 5.67125 15.1614 5.6925C15.0495 5.78937 14.9807 5.88437 14.8864 5.98437C14.6814 6.1925 14.5114 6.36625 14.3257 6.49187C14.2451 6.53875 14.1264 6.5225 14.0732 6.51937L13.9151 6.63188C13.0307 5.70562 11.8276 5.09062 10.482 4.96438L10.4601 4.9625L10.4501 4.77688C10.3957 4.725 10.3307 4.68125 10.3145 4.56875C10.2964 4.34562 10.327 4.10438 10.362 3.81438C10.3814 3.67875 10.4126 3.56625 10.4189 3.41875C10.4195 3.38562 10.4182 3.33625 10.4182 3.30063C10.4182 3.04563 10.2314 2.83812 10.0014 2.83812C9.77198 2.83812 9.58573 3.04563 9.58573 3.30063L9.58636 3.3125C9.58636 3.34688 9.58448 3.38938 9.58636 3.41938C9.59136 3.56688 9.62323 3.67938 9.64198 3.815C9.67698 4.105 9.70698 4.34562 9.68886 4.57C9.66386 4.65562 9.61698 4.72875 9.55448 4.78438L9.55386 4.785L9.54386 4.96062C8.16823 5.07438 6.95511 5.6925 6.07448 6.62687L6.07198 6.62937C6.00011 6.58 5.95136 6.545 5.90261 6.50937L5.92198 6.52312C5.84698 6.53312 5.77198 6.55625 5.67448 6.49875C5.48886 6.37375 5.31886 6.20062 5.11386 5.99187C5.01948 5.89187 4.95136 5.79687 4.83948 5.70125C4.81448 5.67937 4.77511 5.64937 4.74698 5.62812C4.66823 5.56375 4.56761 5.52313 4.45823 5.51813H4.45698C4.45073 5.5175 4.44323 5.5175 4.43573 5.5175C4.31073 5.5175 4.19948 5.57438 4.12573 5.66375L4.12511 5.66437C3.98198 5.84437 4.02761 6.11937 4.22761 6.27875L4.23323 6.28312L4.32011 6.3525C4.43823 6.44 4.54698 6.485 4.66511 6.55438C4.91448 6.70875 5.12011 6.83625 5.28448 6.98937C5.34761 7.0575 5.35948 7.1775 5.36761 7.22937L5.50073 7.34875C4.93386 8.18813 4.59573 9.22312 4.59573 10.3369C4.59573 10.6075 4.61573 10.8737 4.65448 11.1337L4.65073 11.1044L4.47761 11.1544C4.43198 11.2144 4.36698 11.3075 4.29823 11.335C4.08386 11.4025 3.84323 11.4269 3.55261 11.4575C3.41573 11.4694 3.29823 11.4625 3.15261 11.49C3.12198 11.4956 3.07761 11.5069 3.04198 11.515L3.03886 11.5169L3.03323 11.5188C2.78761 11.5781 2.63011 11.8037 2.68073 12.0256C2.73136 12.2481 2.97136 12.3831 3.21823 12.33L3.22386 12.3294L3.23198 12.3269L3.33948 12.3025C3.48136 12.2644 3.58448 12.2081 3.71261 12.1594C3.98761 12.0612 4.21573 11.9788 4.43761 11.9462C4.53073 11.9388 4.62948 12.0037 4.67761 12.0306L4.85823 12C5.28073 13.28 6.13011 14.3225 7.23386 14.9831L7.25823 14.9969L7.18323 15.1787C7.21073 15.2487 7.24073 15.3444 7.22011 15.4137C7.13948 15.6237 7.00073 15.8444 6.84323 16.0912C6.76761 16.2044 6.68886 16.2931 6.62011 16.4237C6.60323 16.4544 6.58261 16.5031 6.56698 16.5356C6.46011 16.765 6.53886 17.0281 6.74448 17.1275C6.95136 17.2275 7.20761 17.1219 7.31948 16.8925V16.8906C7.33636 16.8581 7.35761 16.8156 7.37136 16.785C7.42948 16.65 7.44948 16.5344 7.49136 16.4031C7.60136 16.1263 7.66198 15.8362 7.81386 15.6556C7.85573 15.6056 7.92198 15.5875 7.99323 15.5681L8.08761 15.3975C8.66136 15.625 9.32573 15.7569 10.0214 15.7569C10.707 15.7569 11.3626 15.6287 11.9651 15.395L11.9282 15.4075L12.0164 15.5675C12.0882 15.5906 12.1664 15.6025 12.2295 15.6969C12.3426 15.89 12.4201 16.1194 12.5145 16.3969C12.5564 16.5269 12.5764 16.6425 12.6351 16.7775C12.6482 16.8081 12.6707 16.8525 12.687 16.885C12.7976 17.115 13.0551 17.22 13.262 17.12C13.4676 17.0219 13.5464 16.7575 13.4395 16.5281C13.4226 16.4956 13.402 16.4481 13.3851 16.4163C13.3157 16.2863 13.2376 16.1988 13.162 16.0844C13.0039 15.8375 12.8739 15.6337 12.7926 15.4237C12.7595 15.3156 12.7982 15.2487 12.8245 15.1787C12.8095 15.1606 12.7751 15.0588 12.7551 15.0106C13.887 14.3313 14.737 13.2806 15.1445 12.0306L15.1551 11.9925C15.2082 12.0006 15.302 12.0175 15.3326 12.0244C15.3951 11.9825 15.4526 11.9294 15.5657 11.9375C15.7876 11.97 16.0157 12.0525 16.2907 12.1506C16.4189 12.2006 16.5214 12.2575 16.6639 12.295C16.6939 12.3031 16.737 12.3106 16.772 12.3181L16.7795 12.3206L16.7851 12.3212C17.0326 12.3744 17.272 12.2394 17.3226 12.0169C17.3726 11.795 17.2157 11.5694 16.9701 11.51L16.9695 11.5075ZM13.667 8.08437L12.042 9.53937V9.54375C11.9645 9.61125 11.9157 9.71062 11.9157 9.82063C11.9157 9.98875 12.0289 10.1306 12.1839 10.1737L12.1864 10.1744L12.1889 10.1825L14.2939 10.7894C14.307 10.6625 14.3145 10.5156 14.3145 10.3669C14.3145 10.0219 14.2745 9.68625 14.1982 9.36437L14.2039 9.39375C14.0882 8.8975 13.9032 8.46 13.6564 8.06437L13.667 8.0825V8.08437ZM10.3226 12.5219C10.2601 12.4056 10.1395 12.3281 10.0007 12.3281C9.99573 12.3281 9.99011 12.3281 9.98511 12.3287H9.98573C9.85136 12.3337 9.73636 12.4106 9.67698 12.5212L9.67573 12.5231H9.67386L8.61698 14.4331C9.03136 14.5794 9.50886 14.6637 10.0064 14.6637C10.5014 14.6637 10.9764 14.58 11.4189 14.4269L11.3889 14.4362L10.3307 12.5231H10.3226V12.5219ZM11.8957 11.4444C11.8757 11.4406 11.8526 11.4381 11.8295 11.4381C11.772 11.4381 11.717 11.4513 11.6689 11.475L11.6707 11.4737C11.5464 11.5344 11.462 11.6594 11.462 11.8044C11.462 11.8569 11.4732 11.9075 11.4932 11.9525L11.4926 11.95L11.4901 11.9531L12.332 13.985C13.1282 13.4712 13.7345 12.7212 14.0564 11.8356L14.0657 11.8056L11.8989 11.4387L11.8957 11.4431V11.4444Z" />
    </svg>
  );
}
