import { Component, EventEmitter, Input, OnInit, Output } from '@angular/core';
import { Artifact } from '../../../../../../ng-swagger-gen/models/artifact';
import { Label } from '../../../../../../ng-swagger-gen/models/label';
import { ActivatedRoute, Router } from '@angular/router';
import { Project } from '../../project';
import { artifactDefault } from './artifact';
import { SafeUrl } from '@angular/platform-browser';
import { ArtifactService } from './artifact.service';
import {
    EventService,
    HarborEvent,
} from '../../../../services/event-service/event.service';
import { HG_ICON_NAMES, HG_M_COMMON_ANNOTATION_KEY_PREFIX } from 'src/app/shared/entities/shared.const';
import { ClarityIcons } from '@clr/icons';

@Component({
    selector: 'artifact-summary',
    templateUrl: './artifact-summary.component.html',
    styleUrls: ['./artifact-summary.component.scss'],

    providers: [],
})
export class ArtifactSummaryComponent implements OnInit {
    tagId: string;
    artifactDigest: string;
    sbomDigest?: string;
    activeTab?: string;
    repositoryName: string;
    projectId: string | number;
    referArtifactNameArray: string[] = [];

    labels: Label;
    artifact: Artifact;
    hgLabels: Label[] = [];
    @Output()
    backEvt: EventEmitter<any> = new EventEmitter<any>();
    projectName: string;
    isProxyCacheProject: boolean = false;
    loading: boolean = false;

    constructor(
        private route: ActivatedRoute,
        private router: Router,
        private frontEndArtifactService: ArtifactService,
        private event: EventService
    ) {}

    goBack(): void {
        this.router.navigate([
            'harbor',
            'projects',
            this.projectId,
            'repositories',
            this.repositoryName,
        ]);
    }

    goBackRep(): void {
        this.router.navigate([
            'harbor',
            'projects',
            this.projectId,
            'repositories',
        ]);
    }

    goBackPro(): void {
        this.router.navigate(['harbor', 'projects']);
    }
    jumpDigest(index: number) {
        const arr: string[] = this.referArtifactNameArray.slice(0, index + 1);
        if (arr && arr.length) {
            this.router.navigate([
                'harbor',
                'projects',
                this.projectId,
                'repositories',
                this.repositoryName,
                'artifacts-tab',
                'depth',
                arr.join('-'),
            ]);
        } else {
            this.router.navigate([
                'harbor',
                'projects',
                this.projectId,
                'repositories',
                this.repositoryName,
            ]);
        }
    }

    ngOnInit(): void {
        let depth = this.route.snapshot.params['depth'];
        if (depth) {
            this.referArtifactNameArray = depth.split('-');
        }
        this.repositoryName = this.route.snapshot.params['repo'];
        this.artifactDigest = this.route.snapshot.params['digest'];
        this.projectId = this.route.snapshot.parent.params['id'];
        this.sbomDigest = this.route.snapshot.queryParams['sbomDigest'];
        this.activeTab = this.route.snapshot.queryParams['tab'];
        if (this.repositoryName && this.artifactDigest) {
            const resolverData = this.route.snapshot.data;
            if (resolverData) {
                const pro: Project = <Project>(
                    resolverData['artifactResolver'][1]
                );
                this.projectName = pro.name;
                if (pro.registry_id) {
                    this.isProxyCacheProject = true;
                }
                this.artifact = <Artifact>resolverData['artifactResolver'][0];
                this.getIconFromBackEnd();
                // this.artifact.annotations = {
                //     ...this.artifact.annotations ?? {},
                //     'org.cnai.model.created': '2024-11-03 18:23:10',
                //     'org.cnai.model.authors': 'tom',
                //     'org.cnai.model.url': 'https://huggingface.co/meta-llama/Llama-2-7b-chat-hf/tree/main',
                //     'org.cnai.model.documentation': 'https://huggingface.co/meta-llama/Llama-2-7b-chat-hf/tree/main',
                //     'org.cnai.model.source': 'https://huggingface.co/meta-llama/Llama-2-7b-chat-hf/tree/main',
                //     'org.cnai.model.version': 'v0.0.1',
                //     'org.cnai.model.revision': 'f77b82c1cbbef8423fbddbac20fbe2b7192f13b0',
                //     'org.cnai.model.vendor': 'Alibaba',
                //     'org.cnai.model.licenses': 'llama2',
                //     'org.cnai.model.ref.name': 'meta-llama/Llama-2-7b-chat-hf ',
                //     'org.cnai.model.title': 'Harbor AI Image',
                //     'org.cnai.model.description': 'This is a placeholder for description',
                //     'org.cnai.model.architecture': 'transformer',
                //     'org.cnai.model.family': 'llama3',
                //     'org.cnai.model.name': 'Llama-2-7b-chat-hf',
                //     'org.cnai.model.format': 'pytorch',
                //     'org.cnai.model.param.size': '101200000',
                //     'org.cnai.model.precision': 'bf16',
                //     'org.cnai.model.quantization': 'awq',
                //     'org.cnai.model.files': '[\"src/package.json\",\"src/app.tsx\",\"src/app.scss\",\"src/app.html\",\"src/login/login.html\",\"src/login/login.tsx\",\"src/login/login.scss\",\"readme.td\",\".npmrc\"]'
                // };
                const tagStr = this.artifact.annotations['org.cnai.model.tags'].replace('[', '').replace(']', '');
                const tags = tagStr.split(',').map(item => item.trim());
                this.hgLabels = [
                    'org.cnai.model.authors',
                    'org.cnai.model.family',
                    ...tags
                ].map(key => {
                    const parts = key.split('.');
                    const value = this.artifact?.annotations?.[key] ?? '';

                    return key.startsWith(HG_M_COMMON_ANNOTATION_KEY_PREFIX)
                      ? { name: `${parts[parts.length - 1]}: ${value}` }
                      : { name: key }
                });
            }
        }
        // scroll to the top for harbor container HTML element
        this.event.publish(HarborEvent.SCROLL_TO_POSITION, 0);
    }
    onBack(): void {
        this.backEvt.emit(this.repositoryName);
    }
    showDefaultIcon(event: any) {
        if (event && event.target) {
            event.target.src = artifactDefault;
        }
    }
    getIcon(icon: string): SafeUrl {
        return this.frontEndArtifactService.getIcon(icon);
    }
    getIconFromBackEnd() {
        if (this.artifact && this.artifact.icon) {
            this.frontEndArtifactService.getIconsFromBackEnd([this.artifact]);
        }
    }

    getHgKeysFromArtifact(): string[] {
        const keys = Object.getOwnPropertyNames(this.artifact?.annotations ?? {});
        return keys.filter(key => key.startsWith(HG_M_COMMON_ANNOTATION_KEY_PREFIX));
    }

    isHgArtifact(): boolean {
        return this.getHgKeysFromArtifact().length > 0;
    }

    getHgKeyValueFromArtifact(key: string): string {
        return (this.artifact?.annotations ?? {})?.[key] ?? '';
    }

    getShape(label?: Label): string {
        const labelFormat = (label?.name ?? '').trim().replace(/ /gi, '-').toLowerCase();
        const finalLabelName = `hg-${labelFormat}`;
        const supportedIcons = HG_ICON_NAMES.filter(name => finalLabelName.includes(name)) ?? [];
        if (label.name.toLocaleLowerCase().indexOf('authors') >= 0) {
            return 'administrator';
        }
        if (label.name.toLocaleLowerCase().indexOf('region') >= 0) {
            return 'cloud';
        }
        if (label.name.toLocaleLowerCase().indexOf('meta') >= 0) {
            return 'bundle';
        }
        if (label.name.toLocaleLowerCase().indexOf('lama') >= 0) {
            return 'blocks-group';
        }
        if (label.name.toLocaleLowerCase().startsWith('en')) {
            return 'world';
        }
        return supportedIcons.length > 0
          ? supportedIcons[0]
          : ClarityIcons.has(labelFormat)
          ? labelFormat
          : 'tag'; 
    }
}