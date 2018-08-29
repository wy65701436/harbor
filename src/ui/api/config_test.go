// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
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

package api

import (
	"fmt"
	"testing"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/ui/config"
	"github.com/stretchr/testify/assert"
)

func TestGetConfig(t *testing.T) {
	fmt.Println("Testing getting configurations")
	assert := assert.New(t)
	apiTest := newHarborAPI()

	//case 1: get configurations without admin role
	code, _, err := apiTest.GetConfig(*testUser)
	if err != nil {
		t.Fatalf("failed to get configurations: %v", err)
	}

	assert.Equal(401, code, "the status code of getting configurations with non-admin user should be 401")

	//case 2: get configurations with admin role
	code, cfg, err := apiTest.GetConfig(*admin)
	if err != nil {
		t.Fatalf("failed to get configurations: %v", err)
	}

	if !assert.Equal(200, code, "the status code of getting configurations with admin user should be 200") {
		return
	}

	mode := cfg[common.AUTHMode].Value.(string)
	assert.Equal(common.DBAuth, mode, fmt.Sprintf("the auth mode should be %s", common.DBAuth))
	ccc, err := config.GetSystemCfg()
	if err != nil {
		t.Logf("failed to get system configurations: %v", err)
	}
	t.Logf("%v", ccc)
}

func TestPutConfig(t *testing.T) {
	fmt.Println("Testing modifying configurations")
	assert := assert.New(t)
	apiTest := newHarborAPI()

	cfg := map[string]interface{}{
		common.TokenExpiration: 60,
	}

	code, err := apiTest.PutConfig(*admin, cfg)
	if err != nil {
		t.Fatalf("failed to get configurations: %v", err)
	}

	if !assert.Equal(200, code, "the status code of modifying configurations with admin user should be 200") {
		return
	}
	ccc, err := config.GetSystemCfg()
	if err != nil {
		t.Logf("failed to get system configurations: %v", err)
	}
	t.Logf("%v", ccc)
}

func TestResetConfig(t *testing.T) {
	fmt.Println("Testing resetting configurations")
	assert := assert.New(t)
	apiTest := newHarborAPI()

	code, err := apiTest.ResetConfig(*admin)
	if err != nil {
		t.Errorf("failed to get configurations: %v", err)
		return
	}

	if !assert.Equal(200, code, "unexpected response code") {
		return
	}

	code, cfgs, err := apiTest.GetConfig(*admin)
	if err != nil {
		t.Errorf("failed to get configurations: %v", err)
		return
	}

	if !assert.Equal(200, code, "unexpected response code") {
		return
	}

	value, ok := cfgs[common.TokenExpiration]
	if !ok {
		t.Errorf("%s not found", common.TokenExpiration)
		return
	}

	assert.Equal(int(value.Value.(float64)), 30, "unexpected 30")

	ccc, err := config.GetSystemCfg()
	if err != nil {
		t.Logf("failed to get system configurations: %v", err)
	}
	t.Logf("%v", ccc)
}

func TestPutConfigMaxLength(t *testing.T) {
	fmt.Println("Testing modifying configurations with max length.")
	assert := assert.New(t)
	apiTest := newHarborAPI()

	// length is 512，expected code: 200
	cfg := map[string]interface{}{
		common.LDAPGroupSearchFilter: "OSWvgTrQJuhiPRZt7eCReNku29vrtMBBD2cZt6jl7LQN4OZQcirqEhS2vCnkW8X1OAHMJxiO1LyEY26j" +
			"YhBEiUFliPKDUt8Q9endowT3H60nJibEnCkSRVjix7QujXKRzlmvxcOK76v1oZAoWeHSwwtv7tZrOk16Jj5LTGYdLOnZd2LIgBniTKmceL" +
			"VY5WOgcpmgQCfI5HLbzWsmAqmFfbsDbadirrEDiXYYfZQ0LnF8s6sD4H13eImgenAumXEsBRH43FT37AbNXNxzlaSs8IQYEdPLaMyKoXFb" +
			"rfa0LPipwXnU7bl54IlWOTXwCwum0JGS4qBiMl6LwKUBle34ObZ9fTLh5dFOVE1GdzrGE0kQ7qUmYjMZafQbSXzV80zTc22aZt3RQa9Gxt" +
			"Dn2VqtgcoKAiZHkEySStiwOJtZpwuplyy1jcM3DcN0R9b8IidYAWOsriqetUBThqb75XIZTXAaRWhHLw4ayROYiaw8dPuLRjeVKhdyznqq" +
			"AKxQGyvm",
	}
	code200, _ := apiTest.PutConfig(*admin, cfg)
	assert.Equal(200, code200, "the status code of modifying configurations with admin user should be 200")

	// length is 1059，expected code: 500
	cfg = map[string]interface{}{
		common.LDAPGroupSearchFilter: "J8GtXaDUNTJv6TewTpoyRYGQBLryurjIHdvotcR1v68bC89iaa9Csa4mwKfyCFpPG8UmjOt03z5gNlmn" +
			"qFKzfQzjVlVnoKhNmpLhzWfFXpuLrfxljH0i560Lq9NErPYS06p73N38yVUD6BGubp229lJGenqcYLhaoUru6vay8ChQvouJtcQ1ai5BQ3" +
			"maBKfyfALfI5ca0ETUkrN1wpAXVD8K27iSACnldqgM3WZ6uxXGXYA8HEOA8tEy8bMbIOwtqHuE1zyslQzDKqPUWfFhE1dZjIyipuC7eksF" +
			"dNaIUvlhwWKiCjoxb5X5lkB2ZNjqX2gRn80dfAPfbQIcEPJxey9rldk1rvjfRxLP9NngS6l1wLDY2qk6pNVC9zyS1yfYeBBj8hOhmvb3vN" +
			"mKYe9IslWNIydRYl3jqmVdmL3RF1MW1GU74EpurUwRzeLYtkeBzegp2ZhZNKDaUJ0OysFNKjtyhgdiL6gv06yEvcs7CVHhxn8W6uTeAFiI" +
			"XfTBipJ8Hj0Sv7ZGQZbxcOe66EHAnTLEe5rst312a8mcQbkftY3oVcCuUhmvw2vMvxb9z4p2D42GSWIDsGtpF5FcwnF3VRZeDQUksTgTJ0" +
			"PEtLlB6YCXOBBBvrCsTsegAMLHznesN2OJPN45e6PKaY53Chsj3i4m0Rf7GC2b6FuiIRi3r4VK1bhjGYgHoPQE1X6UGgCFLmyZKH",
	}
	code500, _ := apiTest.PutConfig(*admin, cfg)
	assert.Equal(500, code500, "the status code of modifying configurations with admin user should be 500")
}
