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
	"time"
	"strings"

	"github.com/goharbor/harbor/src/common/models"
	"github.com/astaxie/beego/orm"
)

// AddOIDCUser adds a oidc user
func AddOIDCUser(meta *models.OIDCUser) (int64, error) {
	now := time.Now()
	meta.CreationTime = now
	meta.UpdateTime = now
	id, err := GetOrmer().Insert(meta)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			return 0, ErrDupRows
		}
		return 0, err
	}
	return id, nil
}

// GetOIDCUserByID ...
func GetOIDCUserByID(id int64) (*models.OIDCUser, error) {
	oidcUser := &models.OIDCUser{
		ID: id,
	}
	if err := GetOrmer().Read(oidcUser); err != nil {
		if err == orm.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return oidcUser, nil
}

// GetOIDCUserByID ...
func GetOIDCUserByUserID(id int) (*models.OIDCUser, error) {
	oidcUser := &models.OIDCUser{
		UserID: id,
	}
	if err := GetOrmer().Read(oidcUser); err != nil {
		if err == orm.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return oidcUser, nil
}

// UpdateOIDCUser ...
func UpdateOIDCUser(oidcUser *models.OIDCUser) error {
	oidcUser.UpdateTime = time.Now()
	_, err := GetOrmer().Update(oidcUser)
	return err
}

// DeleteOIDCUser ...
func DeleteOIDCUser(id int64) error {
	_, err := GetOrmer().QueryTable(&models.OIDCUser{}).Filter("ID", id).Delete()
	return err
}
