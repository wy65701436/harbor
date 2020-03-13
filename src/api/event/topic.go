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

package event

import (
	"fmt"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/pkg/artifact"
	"github.com/goharbor/harbor/src/pkg/audit/model"
	v1 "github.com/goharbor/harbor/src/pkg/scan/rest/v1"
	"time"
)

// the event consumers can refer to this file to find all topics and the corresponding event structures

// const definition
const (
	TopicCreateProject     = "CREATE_PROJECT"
	TopicDeleteProject     = "DELETE_PROJECT"
	TopicDeleteRepository  = "DELETE_REPOSITORY"
	TopicPushArtifact      = "PUSH_ARTIFACT"
	TopicPullArtifact      = "PULL_ARTIFACT"
	TopicDeleteArtifact    = "DELETE_ARTIFACT"
	TopicCreateTag         = "CREATE_TAG"
	TopicDeleteTag         = "DELETE_TAG"
	TopicScanningFailed    = "SCANNING_FAILED"
	TopicScanningCompleted = "SCANNING_COMPLETED"
	// QuotaExceedTopic is topic for quota warning event, the usage reaches the warning bar of limitation, like 85%
	TopicQuotaWarning  = "QUOTA_WARNNING"
	TopicQuotaExceed   = "QUOTA_EXCEED"
	TopicUploadChart   = "UPLOAD_CHART"
	TopicDownloadChart = "DOWNLOAD_CHART"
	TopicDeleteChart   = "DELETE_CHART"
)

// CreateProjectEvent is the creating project event
type CreateProjectEvent struct {
	ProjectID int64
	Project   string
	Operator  string
	OccurAt   time.Time
}

// ResolveToAuditLog ...
func (c *CreateProjectEvent) ResolveToAuditLog() (*model.AuditLog, error) {
	auditLog := &model.AuditLog{
		ProjectID:    c.ProjectID,
		OpTime:       c.OccurAt,
		Operation:    "create",
		Username:     c.Operator,
		ResourceType: "project",
		Resource:     fmt.Sprintf("%s", c.Project)}
	return auditLog, nil
}

// DeleteProjectEvent is the deleting project event
type DeleteProjectEvent struct {
	ProjectID int64
	Project   string
	Operator  string
	OccurAt   time.Time
}

// ResolveToAuditLog ...
func (d *DeleteProjectEvent) ResolveToAuditLog() (*model.AuditLog, error) {
	auditLog := &model.AuditLog{
		ProjectID:    d.ProjectID,
		OpTime:       d.OccurAt,
		Operation:    "delete",
		Username:     d.Operator,
		ResourceType: "project",
		Resource:     fmt.Sprintf("%s", d.Project)}
	return auditLog, nil
}

// DeleteRepositoryEvent is the deleting repository event
type DeleteRepositoryEvent struct {
	ProjectID  int64
	Repository string
	Operator   string
	OccurAt    time.Time
}

// ResolveToAuditLog ...
func (d *DeleteRepositoryEvent) ResolveToAuditLog() (*model.AuditLog, error) {
	auditLog := &model.AuditLog{
		ProjectID:    d.ProjectID,
		OpTime:       d.OccurAt,
		Operation:    "delete",
		Username:     d.Operator,
		ResourceType: "project",
		Resource:     d.Repository,
	}
	return auditLog, nil
}

// ArtifactEvent is the pushing/pulling artifact event
type ArtifactEvent struct {
	EventType  string
	Repository string
	Artifact   *artifact.Artifact
	Tag        string // when the artifact is pushed by digest, the tag here will be null
	Operator   string
	OccurAt    time.Time
}

// PushArtifactEvent is the pushing artifact event
type PushArtifactEvent struct {
	*ArtifactEvent
}

// ResolveToAuditLog ...
func (p *PushArtifactEvent) ResolveToAuditLog() (*model.AuditLog, error) {
	auditLog := &model.AuditLog{
		ProjectID:    p.Artifact.ProjectID,
		OpTime:       p.OccurAt,
		Operation:    "create",
		Username:     p.Operator,
		ResourceType: "artifact",
		Resource: fmt.Sprintf("%s:%s",
			p.Artifact.RepositoryName, p.Tag)}
	return auditLog, nil
}

// PullArtifactEvent is the pulling artifact event
type PullArtifactEvent struct {
	*ArtifactEvent
}

// ResolveToAuditLog ...
func (p *PullArtifactEvent) ResolveToAuditLog() (*model.AuditLog, error) {
	auditLog := &model.AuditLog{
		ProjectID:    p.Artifact.ProjectID,
		OpTime:       p.OccurAt,
		Operation:    "pull",
		Username:     p.Operator,
		ResourceType: "artifact",
		Resource: fmt.Sprintf("%s:%s",
			p.Artifact.RepositoryName, p.Tag)}
	return auditLog, nil
}

// DeleteArtifactEvent is the deleting artifact event
type DeleteArtifactEvent struct {
	EventType  string
	Repository string
	Artifact   *artifact.Artifact
	Tags       []string // all the tags that attached to the deleted artifact
	Operator   string
	OccurAt    time.Time
}

// ResolveToAuditLog ...
func (p *DeleteArtifactEvent) ResolveToAuditLog() (*model.AuditLog, error) {
	auditLog := &model.AuditLog{
		ProjectID:    p.Artifact.ProjectID,
		OpTime:       p.OccurAt,
		Operation:    "delete",
		Username:     p.Operator,
		ResourceType: "artifact",
		Resource:     fmt.Sprintf("%s", p.Artifact.RepositoryName)}
	return auditLog, nil
}

// CreateTagEvent is the creating tag event
type CreateTagEvent struct {
	Repository       string
	Tag              string
	AttachedArtifact *artifact.Artifact
	Operator         string
	OccurAt          time.Time
}

// ResolveToAuditLog ...
func (c *CreateTagEvent) ResolveToAuditLog() (*model.AuditLog, error) {
	auditLog := &model.AuditLog{
		ProjectID:    c.AttachedArtifact.ProjectID,
		OpTime:       c.OccurAt,
		Operation:    "create",
		Username:     c.Operator,
		ResourceType: "tag",
		Resource:     fmt.Sprintf("%s:%s", c.Repository, c.Tag)}
	return auditLog, nil
}

// DeleteTagEvent is the deleting tag event
type DeleteTagEvent struct {
	Repository       string
	Tag              string
	AttachedArtifact *artifact.Artifact
	Operator         string
	OccurAt          time.Time
}

// ResolveToAuditLog ...
func (d *DeleteTagEvent) ResolveToAuditLog() (*model.AuditLog, error) {
	auditLog := &model.AuditLog{
		ProjectID:    d.AttachedArtifact.ProjectID,
		OpTime:       d.OccurAt,
		Operation:    "delete",
		Username:     d.Operator,
		ResourceType: "tag",
		Resource:     fmt.Sprintf("%s:%s", d.Repository, d.Tag)}
	return auditLog, nil
}

// ScanImageEvent is scanning image related event data to publish
type ScanImageEvent struct {
	EventType string
	Artifact  *v1.Artifact
	OccurAt   time.Time
	Operator  string
}

// ChartEvent is chart related event data to publish
type ChartEvent struct {
	EventType   string
	ProjectName string
	ChartName   string
	Versions    []string
	OccurAt     time.Time
	Operator    string
}

// QuotaEvent is project quota related event data to publish
type QuotaEvent struct {
	EventType string
	Project   *models.Project
	Resource  *ImgResource
	OccurAt   time.Time
	RepoName  string
	Msg       string
}

// ImgResource include image digest and tag
type ImgResource struct {
	Digest string
	Tag    string
}
