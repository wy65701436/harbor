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
	"strings"
	"time"

	"github.com/astaxie/beego/orm"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils/log"
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

// GetOIDCUserByUserID ...
func GetOIDCUserByUserID(userID int) (*models.OIDCUser, error) {
	var oidcUsers []models.OIDCUser
	n, err := GetOrmer().Raw(`select * from oidc_user where user_id = ? `, userID).QueryRows(&oidcUsers)
	if err != nil {
		return nil, err
	}

	if n == 0 {
		return nil, nil
	}

	return &oidcUsers[0], nil
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

// OnBoardOIDCUser onboard OIDC user, insert use to harbor_user and sub into oidc_user_metadata
func OnBoardOIDCUser(username, email, sub, secret string) error {
	u, err := GetUser(models.User{
		Username: username,
	})
	log.Debugf("Check if user %s exist", username)
	if err != nil {
		return err
	}

	// If user exists, then check user's sub
	// If user doesn't exist, then onboard user.
	if u != nil {
		oidcUser, err := GetOIDCUserByUserID(u.UserID)
		if err != nil {
			return err
		}

		if oidcUser.Sub != sub {
			err := UpdateOIDCUser(&models.OIDCUser{
				ID:     oidcUser.ID,
				UserID: u.UserID,
				Sub:    sub,
			})
			if err != nil {
				return err
			}
		}
	} else {
		o := GetOrmer()
		err := o.Begin()
		if err != nil {
			return err
		}
		user := models.User{
			Username: username,
			Email:    email,
		}
		userID, err := o.Insert(&user)
		if err != nil {
			log.Error(fmt.Errorf("fail to insert user, %v", err))
			o.Rollback()
		}
		oidcUser := models.OIDCUser{
			UserID: int(userID),
			Sub:    sub,
			Secret: secret,
		}
		_, err = o.Insert(&oidcUser)
		if err != nil {
			log.Error(fmt.Errorf("fail to insert oidc user meta, %v", err))
			o.Rollback()
		}
		o.Commit()
	}

	return nil
}
