import { Component, Input, OnInit } from '@angular/core';
import { AdditionsService } from '../additions.service';
import { ErrorHandler } from '../../../../../../shared/units/error-handler';
import { ArtifactTreeItem, ArtifactTreeItemType, initRootArtifactTreeItem, ModelFileOverview, sortArtifactTreeItems } from '../../artifact-file-tree-item';

@Component({
    selector: 'hbr-artifact-files',
    templateUrl: './artifact-files.component.html',
    styleUrls: ['./artifact-files.component.scss'],
})
export class ArtifactFilesComponent implements OnInit {
    @Input()
    modelFileOverviews: ModelFileOverview[] = [];
    rootArtifactTreeItem: ArtifactTreeItem = <ArtifactTreeItem> {
        isRootFolder: true,
        name: 'root',
        file_type: ArtifactTreeItemType.FOLDER,
        children: []
    };
    selectedArtifactTreeItem: ArtifactTreeItem;
    loading: Boolean = false;
    constructor(
        private errorHandler: ErrorHandler,
        private additionsService: AdditionsService
    ) {}

    ngOnInit(): void {
        this.loading = true;
        initRootArtifactTreeItem(this.modelFileOverviews, this.rootArtifactTreeItem);
        this.selectedArtifactTreeItem = this.rootArtifactTreeItem;
        initRootArtifactTreeItem(this.modelFileOverviews, this.rootArtifactTreeItem);
        this.selectedArtifactTreeItem = this.rootArtifactTreeItem;
        this.loading = false;
    }

    openFolder(selectedItem: ArtifactTreeItem): void {
        this.selectedArtifactTreeItem =
          selectedItem.isBackwardFolder
            ? selectedItem.parent
            : selectedItem;
        this.selectedArtifactTreeItem.children =
            sortArtifactTreeItems(this.selectedArtifactTreeItem?.children ?? []);
    }

}