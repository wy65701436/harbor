import { Component, Input, OnChanges, SimpleChanges } from '@angular/core';
import { DatePipe } from '@angular/common';
import { Artifact } from '../../../../../../../ng-swagger-gen/models/artifact';
import { HG_M_COMMON_ANNOTATION_KEY_PREFIX } from 'src/app/shared/entities/shared.const';

enum Types {
    CREATED = 'created',
    TYPE = 'type',
    MEDIA_TYPE = 'media_type',
    MANIFEST_MEDIA_TYPE = 'manifest_media_type',
    DIGEST = 'digest',
    SIZE = 'size',
    PUSH_TIME = 'push_time',
    PULL_TIME = 'pull_time',
}

@Component({
    selector: 'artifact-common-properties',
    templateUrl: './artifact-common-properties.component.html',
    styleUrls: ['./artifact-common-properties.component.scss'],
})
export class ArtifactCommonPropertiesComponent implements OnChanges {
    @Input() artifactDetails: Artifact;
    commonProperties: { [key: string]: any } = {};

    constructor() {}

    ngOnChanges(changes: SimpleChanges) {
        if (changes && changes['artifactDetails']) {
            if (this.artifactDetails) {
                Object.assign(
                    this.commonProperties,
                    this.artifactDetails.extra_attrs,
                    this.artifactDetails.annotations
                );
                for (let name in this.commonProperties) {
                    if (this.commonProperties.hasOwnProperty(name)) {
                        if (typeof this.commonProperties[name] === 'object') {
                            if (this.commonProperties[name] === null) {
                                this.commonProperties[name] = '';
                            } else {
                                this.commonProperties[name] = JSON.stringify(
                                    this.commonProperties[name]
                                );
                            }
                        }
                        if (name === Types.CREATED) {
                            this.commonProperties[name] = new DatePipe(
                                'en-us'
                            ).transform(this.commonProperties[name], 'short');
                        }
                    }
                }
            }
        }
        for (let name in this.commonProperties) {
            if (name.startsWith(HG_M_COMMON_ANNOTATION_KEY_PREFIX) &&
            [
                'org.cnai.model.title',
                'org.cnai.model.url',
                'org.cnai.model.created',
                'org.cnai.model.revision'
            ].indexOf(name) < 0) {
                delete this.commonProperties[name];
            }
        }
    }

    hasCommonProperties(): boolean {
        return JSON.stringify(this.commonProperties) !== '{}';
    }
}
