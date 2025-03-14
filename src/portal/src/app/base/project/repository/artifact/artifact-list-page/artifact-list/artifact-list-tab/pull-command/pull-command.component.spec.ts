// Copyright Project Harbor Authors
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
import { PullCommandComponent } from './pull-command.component';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { SharedTestingModule } from '../../../../../../../../shared/shared.module';
import { ArtifactType } from '../../../../artifact'; // Import the necessary type

describe('PullCommandComponent', () => {
    let component: PullCommandComponent;
    let fixture: ComponentFixture<PullCommandComponent>;

    beforeEach(async () => {
        await TestBed.configureTestingModule({
            declarations: [PullCommandComponent],
            imports: [SharedTestingModule],
        }).compileComponents();

        fixture = TestBed.createComponent(PullCommandComponent);
        component = fixture.componentInstance;

        // Mock the artifact input with a valid value
        component.artifact = {
            type: ArtifactType.IMAGE,
            digest: 'sampleDigest',
            tags: [{ name: 'latest' }],
        };

        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
