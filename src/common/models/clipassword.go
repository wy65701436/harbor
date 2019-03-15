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

package models

import (
	"time"
)

// ClipasswordTable is the name of table in DB that holds the robot object
const ClipasswordTable = "clipassword"

// Clipassword holds the details of a clipassword.
type Clipassword struct {
	ID           int64     `orm:"pk;auto;column(id)" json:"id"`
	UserID       int       `orm:"column(user_id)" json:"user_id"`
	Password     string    `orm:"column(password)" json:"password"`
	Salt         string    `orm:"column(salt)" json:"-"`
	CreationTime time.Time `orm:"column(creation_time);auto_now_add" json:"creation_time"`
}

// TableName ...
func (cli *Clipassword) TableName() string {
	return ClipasswordTable
}