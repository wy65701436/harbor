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

package dao

import (
	"fmt"
	"time"

	"github.com/goharbor/harbor/src/common/models"
)

// AddOIDCUserMetadata adds metadata for a oidc user
func AddOIDCUserMetadata(meta *models.OIDCUserMetaData) error {
	now := time.Now()
	sql := `insert into oidc_user_metadata
				(user_id, name, value, creation_time, update_time)
				 values (?, ?, ?, ?, ?)`
	_, err := GetOrmer().Raw(sql, meta.UserID, meta.Name, meta.Value,
		now, now).Exec()
	return err
}

// UpdateOIDCUserMetadata updates metadata of a oidc user
func UpdateOIDCUserMetadata(meta *models.OIDCUserMetaData) error {
	sql := `update oidc_user_metadata 
				set value = ?, update_time = ? 
				where user_id = ? and name = ?`
	_, err := GetOrmer().Raw(sql, meta.Value, time.Now(), meta.UserID,
		meta.Name).Exec()
	return err
}

// GetOIDCUserMetadata returns the metadata of a oidc user.
func GetOIDCUserMetadata(userID int, name ...string) ([]*models.OIDCUserMetaData, error) {
	oidcUserMetas := []*models.OIDCUserMetaData{}
	params := make([]interface{}, 1)

	sql := `select * from oidc_user_metadata 
				where user_id = ?`
	params = append(params, userID)

	if len(name) > 0 {
		sql += fmt.Sprintf(` and name in ( %s )`, paramPlaceholder(len(name)))
		params = append(params, name)
	}

	_, err := GetOrmer().Raw(sql, params).QueryRows(&oidcUserMetas)
	return oidcUserMetas, err
}

// ListOIDCUserMetadata ...
func ListOIDCUserMetadata(name, value string) ([]*models.OIDCUserMetaData, error) {
	sql := `select * from oidc_user_metadata 
				where name = ? and value = ?`
	metadatas := []*models.OIDCUserMetaData{}
	_, err := GetOrmer().Raw(sql, name, value).QueryRows(&metadatas)
	return metadatas, err
}
