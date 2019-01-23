package token

import (
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/common/utils/test"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/stretchr/testify/assert"
	"os"
	"path"
	"runtime"
	"testing"
)

func TestMain(m *testing.M) {
	_, f, _, ok := runtime.Caller(0)
	if !ok {
		panic("Failed to get current directory")
	}
	keyPath := path.Join(path.Dir(f), "test/private_key.pem")
	os.Setenv("TOKEN_PRIVATE_KEY_PATH", keyPath)

	server, err := test.NewAdminserver(nil)
	if err != nil {
		panic(err)
	}
	defer server.Close()

	if err := os.Setenv("ADMINSERVER_URL", server.URL); err != nil {
		panic(err)
	}
	if err := config.Init(); err != nil {
		panic(err)
	}

	result := m.Run()
	if result != 0 {
		os.Exit(result)
	}
}

func TestValid(t *testing.T) {

	rbacPolicy := &rbac.Policy{
		Resource: "/project/libray/repository",
		Action:   "pull",
	}
	policies := []*rbac.Policy{}
	policies = append(policies, rbacPolicy)

	rClaims := &RobotClaims{
		TokenID:   1,
		ProjectID: 2,
		Policy:    policies,
	}
	assert.Nil(t, rClaims.Valid())
}

func TestUnValidTokenID(t *testing.T) {

	rbacPolicy := &rbac.Policy{
		Resource: "/project/libray/repository",
		Action:   "pull",
	}
	policies := []*rbac.Policy{}
	policies = append(policies, rbacPolicy)

	rClaims := &RobotClaims{
		TokenID:   -1,
		ProjectID: 2,
		Policy:    policies,
	}
	assert.NotNil(t, rClaims.Valid())
}

func TestUnValidProjectID(t *testing.T) {

	rbacPolicy := &rbac.Policy{
		Resource: "/project/libray/repository",
		Action:   "pull",
	}
	policies := []*rbac.Policy{}
	policies = append(policies, rbacPolicy)

	rClaims := &RobotClaims{
		TokenID:   1,
		ProjectID: -2,
		Policy:    policies,
	}
	assert.NotNil(t, rClaims.Valid())
}

func TestUnValidPolicy(t *testing.T) {

	rClaims := &RobotClaims{
		TokenID:   1,
		ProjectID: 2,
		Policy:    nil,
	}
	assert.NotNil(t, rClaims.Valid())
}
