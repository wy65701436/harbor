// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
import { NgModule } from '@angular/core';
import { RouterModule } from '@angular/router';
import { TranslateModule, TranslateStore } from '@ngx-translate/core';
import { FormsModule, ReactiveFormsModule } from '@angular/forms';
import { DateValidatorDirective } from './directives/date-validator.directive';
import { PortValidatorDirective } from './directives/port.directive';
import { MaxLengthExtValidatorDirective } from './directives/max-length-ext.directive';
import { ErrorHandler } from './units/error-handler';
import { ClarityIconsApi } from '@clr/icons/clr-icons-api';
import { ClarityModule } from '@clr/angular';
import { MarkdownModule } from 'ngx-markdown';
import { CommonModule } from '@angular/common';
import { ClipboardModule } from './components/third-party/ngx-clipboard';
import { InlineAlertComponent } from './components/inline-alert/inline-alert.component';
import { NewUserFormComponent } from './components/new-user-form/new-user-form.component';
import { MessageComponent } from './components/global-message/message.component';
import { NavigatorComponent } from './components/navigator/navigator.component';
import { SearchResultComponent } from './components/global-search/search-result.component';
import { GlobalSearchComponent } from './components/global-search/global-search.component';
import { AboutDialogComponent } from './components/about-dialog/about-dialog.component';
import {
    LabelDefaultService,
    LabelService,
    ProjectDefaultService,
    ProjectService,
    ReplicationDefaultService,
    ReplicationService,
    ScanningResultDefaultService,
    ScanningResultService,
    SystemInfoDefaultService,
    SystemInfoService,
    UserPermissionDefaultService,
    UserPermissionService,
} from './services';
import { FilterComponent } from './components/filter/filter.component';
import { GaugeComponent } from './components/gauge/gauge.component';
import { ConfirmationDialogComponent } from './components/confirmation-dialog';
import { ListRepositoryROComponent } from './components/list-repository-ro/list-repository-ro.component';
import { OperationComponent } from './components/operation/operation.component';
import { ViewTokenComponent } from './components/view-token/view-token.component';
import { PushImageButtonComponent } from './components/push-image/push-image.component';
import { CopyInputComponent } from './components/push-image/copy-input.component';
import { ListProjectROComponent } from './components/list-project-ro/list-project-ro.component';
import {
    CronScheduleComponent,
    CronTooltipComponent,
} from './components/cron-schedule';
import { LabelComponent } from './components/label/label.component';
import { LabelSignPostComponent } from './components/label/label-signpost/label-signpost.component';
import { LabelPieceComponent } from './components/label/label-piece/label-piece.component';
import { CreateEditLabelComponent } from './components/label/create-edit-label/create-edit-label.component';
import { DatePickerComponent } from './components/datetime-picker/datetime-picker.component';
import {
    EndpointDefaultService,
    EndpointService,
} from './services/endpoint.service';
import { ImageNameInputComponent } from './components/image-name-input/image-name-input.component';
import { MessageHandlerService } from './services/message-handler.service';
import { HttpClientTestingModule } from '@angular/common/http/testing';
import { RouterTestingModule } from '@angular/router/testing';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { HarborDatetimePipe } from './pipes/harbor-datetime.pipe';
import { RemainingTimeComponent } from './components/remaining-time/remaining-time.component';
import { LabelSelectorComponent } from './components/label-selector/label-selector.component';
import { ScrollSectionDirective } from './directives/scroll/scroll-section.directive';
import { ScrollAnchorDirective } from './directives/scroll/scroll-anchor.directive';
import { AppLevelAlertsComponent } from './components/app-level-alerts/app-level-alerts.component';
// import echarts
import * as echarts from 'echarts/core';
import { PieChart } from 'echarts/charts';
import {
    TitleComponent,
    TooltipComponent,
    GridComponent,
    DatasetComponent,
    TransformComponent,
    LegendComponent,
} from 'echarts/components';
import { LabelLayout, UniversalTransition } from 'echarts/features';
import { CanvasRenderer } from 'echarts/renderers';
import { RobotPermissionsPanelComponent } from './components/robot-permissions-panel/robot-permissions-panel.component';

// register necessary components
echarts.use([
    TitleComponent,
    TooltipComponent,
    GridComponent,
    DatasetComponent,
    TransformComponent,
    PieChart,
    LabelLayout,
    UniversalTransition,
    CanvasRenderer,
    LegendComponent,
]);

// ClarityIcons is publicly accessible from the browser's window object.
declare const ClarityIcons: ClarityIconsApi;

// Add custom icons to ClarityIcons
// Add robot head icon
ClarityIcons.add({
    'robot-head': `
<svg xmlns="http://www.w3.org/2000/svg" width="36" height="36" viewBox="0 0 36 36">
<defs><style>.cls-1{fill:none;}</style></defs><g id="Layer_2" data-name="Layer 2">
<circle cx="12.62" cy="18.6" r="1.5"/><circle cx="23.5" cy="18.5" r="1.5"/>
<path d="M22,28H14a1,1,0,0,1,0-2h8a1,1,0,0,1,0,2Z"/>
<path d="M35,25.22a1,1,0,0,1-1-1V19.38a1,1,0,1,1,2,0v4.84A1,1,0,0,1,35,25.22Z"/>
<path d="M1,25a1,1,0,0,1-1-1V19a1,1,0,0,1,2,0v5A1,1,0,0,1,1,25Z"/>
<path d="M19,8.26A3.26,3.26,0,1,1,22.26,5,3.26,3.26,0,0,1,19,8.26Zm0-4.92A1.66,1.66,0,1,0,20.66,5,1.67,1.67,0,0,0,19,3.34Z"/>
<path d="M29.1,10.49a1,1,0,0,0-.86-.49H20V7.58H18V12h9.67A19.51,19.51,0,0,1,30,21.42,
19.06,19.06,0,0,1,27,32H9.05A19.06,19.06,0,0,1,6,21.42,
19.51,19.51,0,0,1,8.33,12H16V10H7.76a1,1,0,0,0-.86.49A21.18,
21.18,0,0,0,4,21.42,21,21,0,0,0,7.71,33.58a1,1,0,0,0,.81.42h19a1,1,0,0,0,
.81-.42A21,21,0,0,0,32,21.42,21.18,21.18,0,0,0,29.1,10.49Z"/>
<rect class="cls-1" width="36" height="36"/></g></svg>`,
    sbom: `
<?xml version='1.0' encoding='utf-8'?>
<!-- Generator: imaengine 6.0   -->
<svg xmlns:xlink="http://www.w3.org/1999/xlink" xmlns="http://www.w3.org/2000/svg" viewBox="0,0,512,512" style="enable-background:new 0 0 512 512;" version="1.1">
<defs/>
<g id="layer0">
<g transform="matrix(1 0 0 1 0 0)">
<path d="M341.333,213.333L362.666,213.333L362.666,117.333C362.619,116.239 362.399,115.159 362.015,114.133C361.93,113.866 361.951,113.557 361.844,113.29C361.285,111.937 360.454,110.713 359.401,109.695L252.885,3.136C250.88,1.135 248.166,0.0079999 245.333,0L10.667,0C4.776,0 0,4.776 0,10.667L0,458.667C0,464.558 4.776,469.334 10.667,469.334L224,469.334L224,448L21.333,448L21.333,21.333L234.666,21.333L234.666,117.333C234.666,123.224 239.442,128 245.333,128L341.333,128L341.333,213.333L341.333,213.333ZM256,106.667L256,36.427L326.219,106.667L256,106.667L256,106.667Z" fill="#175975"/>
<path d="M501.333,341.333L480,341.333L480,330.666C479.988,289.429 446.55,256.009 405.312,256.02C402.195,256.021 399.081,256.217 395.989,256.607C358.752,261.151 330.666,294.047 330.666,333.119L330.666,341.332L309.333,341.332C303.442,341.332 298.666,346.108 298.666,351.999L298.666,501.332C298.666,507.223 303.442,511.999 309.333,511.999L501.333,511.999C507.224,511.999 512,507.223 512,501.332L512,352C512,346.109 507.224,341.333 501.333,341.333L501.333,341.333ZM352,333.131C352,304.822 372.021,281.035 398.571,277.792C427.788,274.057 454.502,294.715 458.237,323.932C458.523,326.165 458.666,328.415 458.667,330.666L458.667,341.333L352,341.333L352,333.131L352,333.131ZM490.667,490.667L320,490.667L320,362.667L490.667,362.667L490.667,490.667L490.667,490.667Z" fill="#175975"/>
<path d="M394.667,423.797L394.667,458.666C394.667,464.557 399.443,469.333 405.334,469.333C411.225,469.333 416,464.558 416,458.667L416,423.563C426.103,417.57 429.435,404.522 423.443,394.419C419.605,387.948 412.633,383.986 405.11,384C393.369,383.979 383.834,393.479 383.813,405.22C383.798,412.922 387.951,420.028 394.667,423.797L394.667,423.797Z" fill="#175975"/>
</g>
<text font-size="100" font-family="'MesloLGLDZForPowerline-Bold'" fill="#175975" transform="matrix(1.24404 0 0 1.04972 34.5897 178.002)">
<tspan x="0" y="112" textLength="240.82">
<![CDATA[SBOM]]></tspan>
</text>
</g>
</svg>`,
    'hg-text-generation': `
    <svg class="" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" aria-hidden="true" fill="currentColor" focusable="false" role="img" width="1em" height="1em" preserveAspectRatio="xMidYMid meet" viewBox="0 0 18 18"><path d="M16.2607 8.08202L14.468 6.28928C14.3063 6.12804 14.0873 6.03749 13.859 6.03749C13.6307 6.03749 13.4117 6.12804 13.25 6.28928L5.6375 13.904V16.9125H8.64607L16.2607 9.30002C16.422 9.13836 16.5125 8.91935 16.5125 8.69102C16.5125 8.4627 16.422 8.24369 16.2607 8.08202V8.08202ZM8.1953 15.825H6.725V14.3547L11.858 9.22118L13.3288 10.6915L8.1953 15.825ZM14.0982 9.92262L12.6279 8.45232L13.8606 7.21964L15.3309 8.68994L14.0982 9.92262Z"></path><path d="M6.18125 9.84373H7.26875V6.03748H8.9V4.94998H4.55V6.03748H6.18125V9.84373Z"></path><path d="M4.55 11.475H2.375V2.775H11.075V4.95H12.1625V2.775C12.1625 2.48658 12.0479 2.20997 11.844 2.00602C11.64 1.80208 11.3634 1.6875 11.075 1.6875H2.375C2.08658 1.6875 1.80997 1.80208 1.60602 2.00602C1.40207 2.20997 1.2875 2.48658 1.2875 2.775V11.475C1.2875 11.7634 1.40207 12.04 1.60602 12.244C1.80997 12.4479 2.08658 12.5625 2.375 12.5625H4.55V11.475Z"></path></svg>
    `,
    'hg-transformers': `
    <svg class="text-black inline-block text-sm" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" aria-hidden="true" focusable="false" role="img" preserveAspectRatio="xMidYMid meet" width="1em" height="1em" viewBox="0 0 95 88"><path fill="#fff" d="M94.25 70.08a8.28 8.28 0 0 1-.43 6.46 10.57 10.57 0 0 1-3 3.6 25.18 25.18 0 0 1-5.7 3.2 65.74 65.74 0 0 1-7.56 2.65 46.67 46.67 0 0 1-11.42 1.68c-5.42.05-10.09-1.23-13.4-4.5a40.4 40.4 0 0 1-10.14.03c-3.34 3.25-7.99 4.52-13.39 4.47a46.82 46.82 0 0 1-11.43-1.68 66.37 66.37 0 0 1-7.55-2.65c-2.28-.98-4.17-2-5.68-3.2a10.5 10.5 0 0 1-3.02-3.6c-.99-2-1.18-4.3-.42-6.46a8.54 8.54 0 0 1-.33-5.63c.25-.95.66-1.83 1.18-2.61a8.67 8.67 0 0 1 2.1-8.47 8.23 8.23 0 0 1 2.82-2.07 41.75 41.75 0 1 1 81.3-.12 8.27 8.27 0 0 1 3.11 2.19 8.7 8.7 0 0 1 2.1 8.47c.52.78.93 1.66 1.18 2.61a8.61 8.61 0 0 1-.32 5.63Z"></path><path fill="#FFD21E" d="M47.21 76.5a34.75 34.75 0 1 0 0-69.5 34.75 34.75 0 0 0 0 69.5Z"></path><path fill="#FF9D0B" d="M81.96 41.75a34.75 34.75 0 1 0-69.5 0 34.75 34.75 0 0 0 69.5 0Zm-73.5 0a38.75 38.75 0 1 1 77.5 0 38.75 38.75 0 0 1-77.5 0Z"></path><path fill="#3A3B45" d="M58.5 32.3c1.28.44 1.78 3.06 3.07 2.38a5 5 0 1 0-6.76-2.07c.61 1.15 2.55-.72 3.7-.32ZM34.95 32.3c-1.28.44-1.79 3.06-3.07 2.38a5 5 0 1 1 6.76-2.07c-.61 1.15-2.56-.72-3.7-.32Z"></path><path fill="#FF323D" d="M46.96 56.29c9.83 0 13-8.76 13-13.26 0-2.34-1.57-1.6-4.09-.36-2.33 1.15-5.46 2.74-8.9 2.74-7.19 0-13-6.88-13-2.38s3.16 13.26 13 13.26Z"></path><path fill="#3A3B45" fill-rule="evenodd" d="M39.43 54a8.7 8.7 0 0 1 5.3-4.49c.4-.12.81.57 1.24 1.28.4.68.82 1.37 1.24 1.37.45 0 .9-.68 1.33-1.35.45-.7.89-1.38 1.32-1.25a8.61 8.61 0 0 1 5 4.17c3.73-2.94 5.1-7.74 5.1-10.7 0-2.34-1.57-1.6-4.09-.36l-.14.07c-2.31 1.15-5.39 2.67-8.77 2.67s-6.45-1.52-8.77-2.67c-2.6-1.29-4.23-2.1-4.23.29 0 3.05 1.46 8.06 5.47 10.97Z" clip-rule="evenodd"></path><path fill="#FF9D0B" d="M70.71 37a3.25 3.25 0 1 0 0-6.5 3.25 3.25 0 0 0 0 6.5ZM24.21 37a3.25 3.25 0 1 0 0-6.5 3.25 3.25 0 0 0 0 6.5ZM17.52 48c-1.62 0-3.06.66-4.07 1.87a5.97 5.97 0 0 0-1.33 3.76 7.1 7.1 0 0 0-1.94-.3c-1.55 0-2.95.59-3.94 1.66a5.8 5.8 0 0 0-.8 7 5.3 5.3 0 0 0-1.79 2.82c-.24.9-.48 2.8.8 4.74a5.22 5.22 0 0 0-.37 5.02c1.02 2.32 3.57 4.14 8.52 6.1 3.07 1.22 5.89 2 5.91 2.01a44.33 44.33 0 0 0 10.93 1.6c5.86 0 10.05-1.8 12.46-5.34 3.88-5.69 3.33-10.9-1.7-15.92-2.77-2.78-4.62-6.87-5-7.77-.78-2.66-2.84-5.62-6.25-5.62a5.7 5.7 0 0 0-4.6 2.46c-1-1.26-1.98-2.25-2.86-2.82A7.4 7.4 0 0 0 17.52 48Zm0 4c.51 0 1.14.22 1.82.65 2.14 1.36 6.25 8.43 7.76 11.18.5.92 1.37 1.31 2.14 1.31 1.55 0 2.75-1.53.15-3.48-3.92-2.93-2.55-7.72-.68-8.01.08-.02.17-.02.24-.02 1.7 0 2.45 2.93 2.45 2.93s2.2 5.52 5.98 9.3c3.77 3.77 3.97 6.8 1.22 10.83-1.88 2.75-5.47 3.58-9.16 3.58-3.81 0-7.73-.9-9.92-1.46-.11-.03-13.45-3.8-11.76-7 .28-.54.75-.76 1.34-.76 2.38 0 6.7 3.54 8.57 3.54.41 0 .7-.17.83-.6.79-2.85-12.06-4.05-10.98-8.17.2-.73.71-1.02 1.44-1.02 3.14 0 10.2 5.53 11.68 5.53.11 0 .2-.03.24-.1.74-1.2.33-2.04-4.9-5.2-5.21-3.16-8.88-5.06-6.8-7.33.24-.26.58-.38 1-.38 3.17 0 10.66 6.82 10.66 6.82s2.02 2.1 3.25 2.1c.28 0 .52-.1.68-.38.86-1.46-8.06-8.22-8.56-11.01-.34-1.9.24-2.85 1.31-2.85Z"></path><path fill="#FFD21E" d="M38.6 76.69c2.75-4.04 2.55-7.07-1.22-10.84-3.78-3.77-5.98-9.3-5.98-9.3s-.82-3.2-2.69-2.9c-1.87.3-3.24 5.08.68 8.01 3.91 2.93-.78 4.92-2.29 2.17-1.5-2.75-5.62-9.82-7.76-11.18-2.13-1.35-3.63-.6-3.13 2.2.5 2.79 9.43 9.55 8.56 11-.87 1.47-3.93-1.71-3.93-1.71s-9.57-8.71-11.66-6.44c-2.08 2.27 1.59 4.17 6.8 7.33 5.23 3.16 5.64 4 4.9 5.2-.75 1.2-12.28-8.53-13.36-4.4-1.08 4.11 11.77 5.3 10.98 8.15-.8 2.85-9.06-5.38-10.74-2.18-1.7 3.21 11.65 6.98 11.76 7.01 4.3 1.12 15.25 3.49 19.08-2.12Z"></path><path fill="#FF9D0B" d="M77.4 48c1.62 0 3.07.66 4.07 1.87a5.97 5.97 0 0 1 1.33 3.76 7.1 7.1 0 0 1 1.95-.3c1.55 0 2.95.59 3.94 1.66a5.8 5.8 0 0 1 .8 7 5.3 5.3 0 0 1 1.78 2.82c.24.9.48 2.8-.8 4.74a5.22 5.22 0 0 1 .37 5.02c-1.02 2.32-3.57 4.14-8.51 6.1-3.08 1.22-5.9 2-5.92 2.01a44.33 44.33 0 0 1-10.93 1.6c-5.86 0-10.05-1.8-12.46-5.34-3.88-5.69-3.33-10.9 1.7-15.92 2.78-2.78 4.63-6.87 5.01-7.77.78-2.66 2.83-5.62 6.24-5.62a5.7 5.7 0 0 1 4.6 2.46c1-1.26 1.98-2.25 2.87-2.82A7.4 7.4 0 0 1 77.4 48Zm0 4c-.51 0-1.13.22-1.82.65-2.13 1.36-6.25 8.43-7.76 11.18a2.43 2.43 0 0 1-2.14 1.31c-1.54 0-2.75-1.53-.14-3.48 3.91-2.93 2.54-7.72.67-8.01a1.54 1.54 0 0 0-.24-.02c-1.7 0-2.45 2.93-2.45 2.93s-2.2 5.52-5.97 9.3c-3.78 3.77-3.98 6.8-1.22 10.83 1.87 2.75 5.47 3.58 9.15 3.58 3.82 0 7.73-.9 9.93-1.46.1-.03 13.45-3.8 11.76-7-.29-.54-.75-.76-1.34-.76-2.38 0-6.71 3.54-8.57 3.54-.42 0-.71-.17-.83-.6-.8-2.85 12.05-4.05 10.97-8.17-.19-.73-.7-1.02-1.44-1.02-3.14 0-10.2 5.53-11.68 5.53-.1 0-.19-.03-.23-.1-.74-1.2-.34-2.04 4.88-5.2 5.23-3.16 8.9-5.06 6.8-7.33-.23-.26-.57-.38-.98-.38-3.18 0-10.67 6.82-10.67 6.82s-2.02 2.1-3.24 2.1a.74.74 0 0 1-.68-.38c-.87-1.46 8.05-8.22 8.55-11.01.34-1.9-.24-2.85-1.31-2.85Z"></path><path fill="#FFD21E" d="M56.33 76.69c-2.75-4.04-2.56-7.07 1.22-10.84 3.77-3.77 5.97-9.3 5.97-9.3s.82-3.2 2.7-2.9c1.86.3 3.23 5.08-.68 8.01-3.92 2.93.78 4.92 2.28 2.17 1.51-2.75 5.63-9.82 7.76-11.18 2.13-1.35 3.64-.6 3.13 2.2-.5 2.79-9.42 9.55-8.55 11 .86 1.47 3.92-1.71 3.92-1.71s9.58-8.71 11.66-6.44c2.08 2.27-1.58 4.17-6.8 7.33-5.23 3.16-5.63 4-4.9 5.2.75 1.2 12.28-8.53 13.36-4.4 1.08 4.11-11.76 5.3-10.97 8.15.8 2.85 9.05-5.38 10.74-2.18 1.69 3.21-11.65 6.98-11.76 7.01-4.31 1.12-15.26 3.49-19.08-2.12Z"></path></svg>
    `,
    'hg-pytorch': `
    <svg class="text-black inline-block text-sm" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" aria-hidden="true" focusable="false" role="img" width="1em" height="1em" preserveAspectRatio="xMidYMid meet" viewBox="0 0 32 32"><defs><clipPath id="icon-pytorch-a"><rect x="3.05" y="0.5" width="25.73" height="31" fill="none"></rect></clipPath></defs><g clip-path="url(#icon-pytorch-a)"><path d="M24.94,9.51a12.81,12.81,0,0,1,0,18.16,12.68,12.68,0,0,1-18,0,12.81,12.81,0,0,1,0-18.16l9-9V5l-.84.83-6,6a9.58,9.58,0,1,0,13.55,0ZM20.44,9a1.68,1.68,0,1,1,1.67-1.67A1.68,1.68,0,0,1,20.44,9Z" fill="#ee4c2c"></path></g></svg>
    `,
    'hg-safetensors': `
    <svg class="text-black inline-block text-sm" viewBox="0 0 57 44" fill="none" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" aria-hidden="true" focusable="false" role="img" width="1em" height="1em" preserveAspectRatio="xMidYMid meet"><path d="M36.816 20.1474L54.9918 27.4409C55.5142 27.6506 55.9623 28.0112 56.2788 28.4766C56.5954 28.9421 56.7661 29.4913 56.7691 30.0542C56.7722 30.6171 56.6074 31.1682 56.2959 31.637C55.9844 32.1059 55.5402 32.4713 55.0201 32.6866L29.953 43.0646C29.2593 43.3518 28.4799 43.3518 27.7862 43.0646L2.71624 32.6894C2.19613 32.4741 1.75197 32.1087 1.44044 31.6398C1.12892 31.171 0.964165 30.62 0.967204 30.057C0.970244 29.4941 1.14094 28.9449 1.45751 28.4794C1.77408 28.014 2.22216 27.6534 2.74456 27.4437L21.2404 20.0227C22.2997 19.5979 25.6477 20.8441 28.8682 20.8555C32.3096 20.8668 35.6292 19.6715 36.816 20.1474ZM11.3042 30.1119L28.8682 37.3828L46.435 30.1119L28.8682 23.0619L11.3042 30.1119ZM29.9247 0.388251L54.9918 10.4462C55.5142 10.6559 55.9623 11.0165 56.2788 11.482C56.5954 11.9474 56.7661 12.4967 56.7691 13.0596C56.7722 13.6225 56.6074 14.1735 56.2959 14.6424C55.9844 15.1112 55.5402 15.4766 55.0201 15.6919L29.953 26.07C29.2593 26.3572 28.4799 26.3572 27.7862 26.07L2.71624 15.6948C2.19613 15.4795 1.75197 15.1141 1.44044 14.6452C1.12892 14.1763 0.964165 13.6253 0.967204 13.0624C0.970244 12.4995 1.14094 11.9503 1.45751 11.4848C1.77408 11.0193 2.22216 10.6588 2.74456 10.4491L27.8117 0.388251C28.4896 0.1157 29.2467 0.1157 29.9247 0.388251ZM11.3042 13.1172L28.8682 20.3881L46.435 13.1172L28.8682 6.06729L11.3042 13.1172Z" fill="currentColor"></path></svg>
    `,
    'hg-english': `
    <svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" aria-hidden="true" role="img" class="text-green-600/80" preserveAspectRatio="xMidYMid meet" width="1em" height="1em" viewBox="0 0 10 10"><path fill-rule="evenodd" clip-rule="evenodd" d="M0.625 5C0.625 6.16032 1.08594 7.27312 1.90641 8.09359C2.72688 8.91406 3.83968 9.375 5 9.375C6.16032 9.375 7.27312 8.91406 8.09359 8.09359C8.91406 7.27312 9.375 6.16032 9.375 5C9.375 3.83968 8.91406 2.72688 8.09359 1.90641C7.27312 1.08594 6.16032 0.625 5 0.625C3.83968 0.625 2.72688 1.08594 1.90641 1.90641C1.08594 2.72688 0.625 3.83968 0.625 5ZM7.64365 7.48027C7.61734 7.50832 7.59054 7.53598 7.56326 7.56326C7.13828 7.98824 6.61864 8.2968 6.0539 8.46842C6.29802 8.11949 6.49498 7.64804 6.63475 7.09483C7.00845 7.18834 7.35014 7.3187 7.64365 7.48027ZM8.10076 6.87776C8.37677 6.42196 8.55005 5.90894 8.60556 5.37499H6.86808C6.85542 5.71597 6.82551 6.04557 6.77971 6.35841C7.25309 6.47355 7.68808 6.6414 8.062 6.85549C8.07497 6.86283 8.08789 6.87025 8.10076 6.87776ZM6.03795 6.22536C6.07708 5.95737 6.1044 5.67232 6.11705 5.37499H3.88295C3.89666 5.69742 3.92764 6.00542 3.9722 6.29287C4.37075 6.21726 4.79213 6.17749 5.224 6.17749C5.50054 6.17749 5.77294 6.19376 6.03795 6.22536ZM4.1261 7.02673C4.34894 7.84835 4.68681 8.375 5 8.375C5.32122 8.375 5.66839 7.82101 5.8908 6.963C5.67389 6.93928 5.45082 6.92699 5.224 6.92699C4.84316 6.92699 4.47332 6.96176 4.1261 7.02673ZM3.39783 7.21853C3.53498 7.71842 3.72038 8.14579 3.9461 8.46842C3.42141 8.30898 2.93566 8.03132 2.52857 7.65192C2.77253 7.48017 3.06711 7.33382 3.39783 7.21853ZM3.23916 6.48077C3.18263 6.13193 3.14625 5.76074 3.13192 5.37499H1.39444C1.4585 5.99112 1.67936 6.57938 2.03393 7.08403C2.3706 6.83531 2.78055 6.63162 3.23916 6.48077ZM1.39444 4.62499H3.13192C3.14615 4.24204 3.18211 3.87344 3.23794 3.52681C2.77814 3.37545 2.36731 3.17096 2.03024 2.92123C1.67783 3.42469 1.45828 4.011 1.39444 4.62499ZM2.5237 2.35262C2.76812 2.52552 3.06373 2.67281 3.39584 2.78875C3.53318 2.28573 3.71928 1.85578 3.9461 1.53158C3.41932 1.69166 2.93178 1.97089 2.5237 2.35262ZM3.97101 3.71489C3.92709 4.00012 3.89654 4.30547 3.88295 4.62499H6.11705C6.10453 4.33057 6.07761 4.04818 6.03909 3.78248C5.77372 3.81417 5.50093 3.83049 5.224 3.83049C4.79169 3.83049 4.3699 3.79065 3.97101 3.71489ZM5.8928 3.04476C5.67527 3.06863 5.45151 3.08099 5.224 3.08099C4.84241 3.08099 4.47186 3.04609 4.12405 2.98086C4.34686 2.1549 4.68584 1.625 5 1.625C5.32218 1.625 5.67048 2.18233 5.8928 3.04476ZM6.78083 3.6493C6.826 3.95984 6.85552 4.28682 6.86808 4.62499H8.60556C8.55029 4.09337 8.37827 3.58251 8.10436 3.1282C8.0903 3.1364 8.07618 3.14449 8.062 3.15249C7.68838 3.36641 7.25378 3.53417 6.78083 3.6493ZM7.64858 2.52499C7.35446 2.68754 7.0117 2.81868 6.63664 2.91268C6.49676 2.35623 6.29913 1.88209 6.0539 1.53158C6.61864 1.7032 7.13828 2.01176 7.56326 2.43674C7.59224 2.46572 7.62068 2.49514 7.64858 2.52499Z" fill="currentColor"></path></svg>
    `,
    'hg-text-generation-inference': `
    <svg class="" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" aria-hidden="true" role="img" width="1em" height="1em" preserveAspectRatio="xMidYMid meet" viewBox="0 0 12 12"><path fill="#23B0FF" d="m9.6 3.6-3.2-2a1 1 0 0 0-1.1 0L2 3.7a1 1 0 0 0-.3 1.6H10a1 1 0 0 0-.3-1.6Z"></path><path fill="#2094FF" d="m6.7 9.7 3.2-4.5-.4-.8H5.7v4.8l1 .5Z"></path><path fill="#6BCAFF" d="M4.9 9.7 1.7 5.2l.4-.8h3.8v4.8l-1 .5Z"></path><path fill="#000" fill-rule="evenodd" d="M9.9 3.2c.8.5 1 1.5.5 2.3L7 10c-.6.9-2 .9-2.6 0L1.3 5.5c-.5-.8-.3-1.8.5-2.3l3.2-2c.5-.3 1.2-.3 1.7 0l3.2 2ZM6.4 5h3l-3 4.2V5ZM5.3 5h-3l3 4.2V5Zm3.8-1L6 2a.5.5 0 0 0-.5 0L2.6 4H9Z" clip-rule="evenodd"></path></svg>
    `,
    'hg-inference-endpoints': `
    <svg class="" width="1em" height="1em" viewBox="0 0 74 75" fill="none" xmlns="http://www.w3.org/2000/svg"><path fill-rule="evenodd" clip-rule="evenodd" d="M46.6517 14.9309C48.2799 11.0495 45.4292 6.76251 41.2202 6.76251H24.1817C19.8066 6.76251 15.9021 9.50789 14.4218 13.6249L8.51785 30.0453C6.77792 34.8844 6.77792 40.1785 8.51785 45.0177L14.4218 61.438C15.9021 65.5551 19.8066 68.3004 24.1817 68.3004H41.2202C45.4292 68.3004 48.2799 64.0134 46.6517 60.132L40.7616 46.0903C38.465 40.6155 38.465 34.4475 40.7616 28.9727L46.6517 14.9309Z" fill="#861FFF" stroke="black" stroke-width="12.6835" stroke-linejoin="round"></path><circle cx="53.1334" cy="37.5315" r="13.9518" fill="#FF3270" stroke="black" stroke-width="12.6835" stroke-linejoin="round"></circle></svg>
    `,
    'hg-license': `
    <svg class="text-xs text-gray-900" width="1em" height="1em" viewBox="0 0 10 10" fill="none" xmlns="http://www.w3.org/2000/svg"><path d="M1.46009 5.0945V6.88125C1.46009 7.25201 1.75937 7.55129 2.13012 7.55129C2.50087 7.55129 2.80016 7.25201 2.80016 6.88125V5.0945C2.80016 4.72375 2.50087 4.42446 2.13012 4.42446C1.75937 4.42446 1.46009 4.72375 1.46009 5.0945ZM4.14022 5.0945V6.88125C4.14022 7.25201 4.4395 7.55129 4.81026 7.55129C5.18101 7.55129 5.48029 7.25201 5.48029 6.88125V5.0945C5.48029 4.72375 5.18101 4.42446 4.81026 4.42446C4.4395 4.42446 4.14022 4.72375 4.14022 5.0945ZM1.23674 9.78473H8.38377C8.75452 9.78473 9.0538 9.48545 9.0538 9.1147C9.0538 8.74395 8.75452 8.44466 8.38377 8.44466H1.23674C0.865993 8.44466 0.566711 8.74395 0.566711 9.1147C0.566711 9.48545 0.865993 9.78473 1.23674 9.78473ZM6.82036 5.0945V6.88125C6.82036 7.25201 7.11964 7.55129 7.49039 7.55129C7.86114 7.55129 8.16042 7.25201 8.16042 6.88125V5.0945C8.16042 4.72375 7.86114 4.42446 7.49039 4.42446C7.11964 4.42446 6.82036 4.72375 6.82036 5.0945ZM4.39484 0.623142L0.865993 2.48137C0.682851 2.57517 0.566711 2.76725 0.566711 2.97273C0.566711 3.28094 0.816857 3.53109 1.12507 3.53109H8.49991C8.80365 3.53109 9.0538 3.28094 9.0538 2.97273C9.0538 2.76725 8.93766 2.57517 8.75452 2.48137L5.22568 0.623142C4.9666 0.484669 4.65391 0.484669 4.39484 0.623142V0.623142Z" fill="currentColor"></path></svg>
    `
});

@NgModule({
    imports: [
        TranslateModule.forChild({
            extend: true,
        }),
        FormsModule,
        CommonModule,
        ClarityModule,
        MarkdownModule.forRoot(),
        RouterModule,
        ReactiveFormsModule,
        ClipboardModule,
    ],
    declarations: [
        MaxLengthExtValidatorDirective,
        PortValidatorDirective,
        DateValidatorDirective,
        ScrollSectionDirective,
        ScrollAnchorDirective,
        InlineAlertComponent,
        NewUserFormComponent,
        MessageComponent,
        NavigatorComponent,
        SearchResultComponent,
        GlobalSearchComponent,
        AboutDialogComponent,
        FilterComponent,
        GaugeComponent,
        ConfirmationDialogComponent,
        ListRepositoryROComponent,
        OperationComponent,
        ViewTokenComponent,
        PushImageButtonComponent,
        CopyInputComponent,
        ListProjectROComponent,
        CronTooltipComponent,
        LabelComponent,
        LabelSignPostComponent,
        LabelPieceComponent,
        CreateEditLabelComponent,
        CronScheduleComponent,
        DatePickerComponent,
        ImageNameInputComponent,
        HarborDatetimePipe,
        RemainingTimeComponent,
        LabelSelectorComponent,
        AppLevelAlertsComponent,
        RobotPermissionsPanelComponent,
    ],
    exports: [
        TranslateModule,
        FormsModule,
        ReactiveFormsModule,
        ClarityModule,
        CommonModule,
        ClipboardModule,
        MarkdownModule,
        MaxLengthExtValidatorDirective,
        PortValidatorDirective,
        DateValidatorDirective,
        ScrollSectionDirective,
        ScrollAnchorDirective,
        InlineAlertComponent,
        NewUserFormComponent,
        MessageComponent,
        NavigatorComponent,
        SearchResultComponent,
        GlobalSearchComponent,
        AboutDialogComponent,
        FilterComponent,
        GaugeComponent,
        ConfirmationDialogComponent,
        ListRepositoryROComponent,
        OperationComponent,
        ViewTokenComponent,
        PushImageButtonComponent,
        CopyInputComponent,
        ListProjectROComponent,
        CronTooltipComponent,
        LabelComponent,
        LabelSignPostComponent,
        LabelPieceComponent,
        CreateEditLabelComponent,
        CronScheduleComponent,
        DatePickerComponent,
        ImageNameInputComponent,
        HarborDatetimePipe,
        RemainingTimeComponent,
        LabelSelectorComponent,
        AppLevelAlertsComponent,
        RobotPermissionsPanelComponent,
    ],
    providers: [
        { provide: EndpointService, useClass: EndpointDefaultService },
        { provide: ReplicationService, useClass: ReplicationDefaultService },
        { provide: LabelService, useClass: LabelDefaultService },
        { provide: SystemInfoService, useClass: SystemInfoDefaultService },
        {
            provide: ScanningResultService,
            useClass: ScanningResultDefaultService,
        },
    ],
})
export class SharedModule {}

// this module is only for testing, you should only import this module in *.spec.ts files
@NgModule({
    imports: [
        BrowserAnimationsModule,
        SharedModule,
        HttpClientTestingModule,
        RouterTestingModule,
    ],
    exports: [
        BrowserAnimationsModule,
        SharedModule,
        HttpClientTestingModule,
        RouterTestingModule,
    ],
    providers: [
        TranslateStore,
        { provide: ProjectService, useClass: ProjectDefaultService },
        { provide: ErrorHandler, useClass: MessageHandlerService },
        {
            provide: UserPermissionService,
            useClass: UserPermissionDefaultService,
        },
    ],
})
export class SharedTestingModule {}
