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

	user111 := &models.User{
		Username: "user111",
		Email:    "user111@email.com",
	}
	user222 := &models.User{
		Username: "user222",
		Email:    "user222@email.com",
	}
	err := OnBoardUser(user111)
	require.Nil(t, err)
	err = OnBoardUser(user222)
	require.Nil(t, err)

	meta1 := &models.OIDCUser{
		UserID: user111.UserID,
		Sub:    "QWE123123RT1",
		Secret: "QWEQWE1",
	}
	meta2 := &models.OIDCUser{
		UserID: user222.UserID,
		Sub:    "QWE123123RT2",
		Secret: "QWEQWE2",
	}

	// test add
	_, err = AddOIDCUser(meta1)
	require.Nil(t, err)
	defer func() {
		// clean up
		err := DeleteOIDCUser(meta1.ID)
		require.Nil(t, err)
	}()
	_, err = AddOIDCUser(meta2)
	require.Nil(t, err)
	defer func() {
		// clean up
		err := DeleteOIDCUser(meta2.ID)
		require.Nil(t, err)
	}()

	// test get
	oidcUser1, err := GetOIDCUserByID(meta1.ID)
	require.Nil(t, err)
	assert.Equal(t, meta1.UserID, oidcUser1.UserID)

	// test update
	meta3 := &models.OIDCUser{
		ID:     meta1.ID,
		UserID: meta1.UserID,
		Sub:    "newSub",
	}
	require.Nil(t, UpdateOIDCUser(meta3))
	oidcUser1Update, err := GetOIDCUserByID(meta1.ID)
	require.Nil(t, err)
	assert.Equal(t, "newSub", oidcUser1Update.Sub)

}
