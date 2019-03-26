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
	"testing"

	"github.com/goharbor/harbor/src/common/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOIDCUserMetaDaoMethods(t *testing.T) {

	meta1 := &models.OIDCUser{
		UserID: 1,
		Sub:    "QWE123123RT",
		Secret: "QWEQWE",
	}
	meta2 := &models.OIDCUser{
		UserID: 2,
		Sub:    "QWE123123RT",
		Secret: "QWEQWE",
	}

	// test add
	_, err := AddOIDCUser(meta1)
	require.Nil(t, err)
	defer func() {
		// clean up
		_, err := GetOrmer().Raw(`delete from oidc_user
			where user_id = 1`).Exec()
		require.Nil(t, err)
	}()
	_, err = AddOIDCUser(meta2)
	require.Nil(t, err)
	defer func() {
		// clean up
		_, err := GetOrmer().Raw(`delete from oidc_user
			where user_id = 2`).Exec()
		require.Nil(t, err)
	}()

	// test get
	oidcUser1, err := GetOIDCUserByID(1)
	require.Nil(t, err)
	assert.Equal(t, 1, oidcUser1.UserID)

	// test update
	userMeta := models.OIDCUser{
		UserID: 1,
		Sub:    "newSub",
	}
	require.Nil(t, UpdateOIDCUser(&userMeta))
	oidcUser1Update, err := GetOIDCUserByID(1)
	require.Nil(t, err)
	assert.Equal(t, "newSub", oidcUser1Update.Sub)

}
