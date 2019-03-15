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
	"github.com/astaxie/beego/orm"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils"
	"time"
)

// AddClipassword ...
func AddClipassword(clipwd *models.Clipassword) (int64, error) {
	o := GetOrmer()
	now := time.Now()
	salt := utils.GenerateRandomString()
	pwd := utils.GenerateRandomString()
	sql := `insert into clipassword (user_id, password, salt, creation_time)
				 values (?, ?, ?, ?) RETURNING id`
	var clipwdID int64
	err := o.Raw(sql, clipwd.UserID, utils.Encrypt(pwd, salt), salt, now).QueryRow(&clipwdID)
	if err != nil {
		return 0, err
	}
	return clipwdID, nil

}

// GetClipasswordByID ...
func GetClipasswordByID(id int64) (*models.Clipassword, error) {
	clipwd := models.Clipassword{}
	_, err := GetOrmer().QueryTable(&models.Clipassword{}).Filter("ID", id).All(&clipwd)
	if err != nil {
		if err == orm.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &clipwd, nil
}

// GetClipasswordByUserId ...
func GetClipasswordByUserId(userID int) (*models.Clipassword, error) {
	o := GetOrmer()
	var clis []models.Clipassword
	n, err := o.Raw(`select * from clipassword where user_id = ? `, userID).QueryRows(&clis)
	if err != nil {
		return nil, err
	}

	if n == 0 {
		return nil, nil
	}

	return &clis[0], nil
}

// DeleteClipasswordById ...
func DeleteClipasswordById(id int64) error {
	_, err := GetOrmer().QueryTable(&models.Clipassword{}).Filter("ID", id).Delete()
	return err
}

// DeleteClipasswordByUserId ...
func DeleteClipasswordByUserId(id int) error {
	_, err := GetOrmer().QueryTable(&models.Clipassword{}).Filter("UserID", id).Delete()
	return err
}