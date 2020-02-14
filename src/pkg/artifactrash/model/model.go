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

package model

import (
	"github.com/astaxie/beego/orm"
)

func init() {
	orm.RegisterModel(&ArtifactTrash{})
}

// ArtifactTrash records the deleted artifact
type ArtifactTrash struct {
	ID                int64  `orm:"pk;auto;column(id)"`
	ArtifactID        int64  `orm:"column(artifact_id)"`
	Type              string `orm:"column(type)"`
	MediaType         string `orm:"column(media_type)"`
	ManifestMediaType string `orm:"column(manifest_media_type)"`
	ProjectID         int64  `orm:"column(project_id)"`
	RepositoryID      int64  `orm:"column(repository_id)"`
	Digest            string `orm:"column(digest)"`
}

// TableName for artifact trash
func (at *ArtifactTrash) TableName() string {
	return "artifact_trash"
}
