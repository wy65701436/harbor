import { ComponentFixture, TestBed } from '@angular/core/testing';
import { NO_ERRORS_SCHEMA } from '@angular/core';
import { AdditionsService } from '../additions.service';
import { ArtifactFilesComponent } from './artifact-files.component';
import { ErrorHandler } from '../../../../../../shared/units/error-handler';
import { SharedTestingModule } from '../../../../../../shared/shared.module';
import { ModelFileOverview } from '../../artifact-file-tree-item';

describe('ArtifactFilesComponent', () => {
    let component: ArtifactFilesComponent;
    let fixture: ComponentFixture<ArtifactFilesComponent>;
    const mockedModelFileOverviews: ModelFileOverview[] = [
        {
            file: 'root/test.txt',
            file_hash: 'xxxxxxxx'
        }
    ];
    beforeEach(async () => {
        await TestBed.configureTestingModule({
            imports: [SharedTestingModule],
            declarations: [ArtifactFilesComponent],
            providers: [
                ErrorHandler,
                { provide: AdditionsService },
            ],
            schemas: [NO_ERRORS_SCHEMA],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(ArtifactFilesComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
    it('should get build history list and render', async () => {
        component.modelFileOverviews = mockedModelFileOverviews;
        component.ngOnInit();
        fixture.detectChanges();
        await fixture.whenStable();
        const rows = fixture.nativeElement.getElementsByClassName('label');
        expect(rows.length).toEqual(2);
    });
});