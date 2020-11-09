package robot

import (
	"context"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/utils"
	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils/test"
	core_cfg "github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/permission/types"
	rbac_model "github.com/goharbor/harbor/src/pkg/rbac/model"
	"github.com/goharbor/harbor/src/pkg/robot2/model"
	"github.com/goharbor/harbor/src/testing/mock"
	"github.com/goharbor/harbor/src/testing/pkg/project"
	"github.com/goharbor/harbor/src/testing/pkg/rbac"
	"github.com/goharbor/harbor/src/testing/pkg/robot2"
	"github.com/stretchr/testify/suite"
	"os"
	"testing"
)

type ControllerTestSuite struct {
	suite.Suite
}

func (suite *ControllerTestSuite) TestGet() {
	projectMgr := &project.Manager{}
	rbacMgr := &rbac.Manager{}
	robotMgr := &robot2.Manager{}

	c := controller{robotMgr: robotMgr, rbacMgr: rbacMgr, proMgr: projectMgr}
	ctx := context.TODO()
	projectMgr.On("Get", mock.Anything, mock.Anything).Return(&models.Project{ProjectID: 1, Name: "library"}, nil)
	robotMgr.On("Get", mock.Anything, mock.Anything).Return(&model.Robot{
		Name:        "test",
		Description: "test get method",
		ProjectID:   1,
		Secret:      utils.RandStringBytes(10),
	}, nil)
	rbacMgr.On("GetPermissionsByRole", mock.Anything, mock.Anything, mock.Anything).Return([]*rbac_model.RolePermissions{
		{
			RoleType: ROBOTTYPE,
			RoleID:   1,
			Scope:    "/project/1",
			Resource: "repository",
			Action:   "pull",
		},
		{
			RoleType: ROBOTTYPE,
			RoleID:   1,
			Scope:    "/project/1",
			Resource: "repository",
			Action:   "push",
		},
	}, nil)
	robot, err := c.Get(ctx, int64(1), &Option{
		WithPermission: true,
	})
	suite.Nil(err)

	suite.Equal("project", robot.Permissions[0].Kind)
	suite.Equal("library", robot.Permissions[0].Namespace)
	suite.Equal("pull", robot.Permissions[0].Access[0].Action.String())
	suite.Equal("project", robot.Level)

}

func (suite *ControllerTestSuite) TestCreate() {
	secretKeyPath := "/tmp/secretkey"
	_, err := test.GenerateKey(secretKeyPath)
	suite.Nil(err)
	defer os.Remove(secretKeyPath)
	os.Setenv("KEY_PATH", secretKeyPath)

	conf := map[string]interface{}{
		common.RobotTokenDuration: "30",
	}
	core_cfg.InitWithSettings(conf)

	projectMgr := &project.Manager{}
	rbacMgr := &rbac.Manager{}
	robotMgr := &robot2.Manager{}

	c := controller{robotMgr: robotMgr, rbacMgr: rbacMgr, proMgr: projectMgr}
	ctx := context.TODO()
	projectMgr.On("Get", mock.Anything, mock.Anything).Return(&models.Project{ProjectID: 1, Name: "library"}, nil)
	robotMgr.On("Create", mock.Anything, mock.Anything).Return(int64(1), nil)
	rbacMgr.On("CreateRbacPolicy", mock.Anything, mock.Anything, mock.Anything).Return(int64(1), nil)
	rbacMgr.On("CreatePermission", mock.Anything, mock.Anything, mock.Anything).Return(int64(1), nil)

	r, err := c.Create(ctx, &Robot{
		Robot: model.Robot{
			Name:        "testcreate",
			Description: "testcreate",
			ExpiresAt:   0,
		},
		ProjectName: "library",
		Level:       LEVELPROJECT,
		Permissions: []*Permission{
			{
				Kind:      "project",
				Namespace: "library",
				Access: []*types.Policy{
					{
						Resource: "repository",
						Action:   "push",
					},
					{
						Resource: "repository",
						Action:   "pull",
					},
				},
			},
		},
	})
	suite.Nil(err)
	suite.Equal("robot$testcreate", r.Name)
}

func (suite *ControllerTestSuite) TestDelete() {
	projectMgr := &project.Manager{}
	rbacMgr := &rbac.Manager{}
	robotMgr := &robot2.Manager{}

	c := controller{robotMgr: robotMgr, rbacMgr: rbacMgr, proMgr: projectMgr}
	ctx := context.TODO()

	robotMgr.On("Delete", mock.Anything, mock.Anything).Return(nil)
	rbacMgr.On("DeletePermissionByRole", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	err := c.Delete(ctx, int64(1))
	suite.Nil(err)
}

func (suite *ControllerTestSuite) TestUpdate() {
	projectMgr := &project.Manager{}
	rbacMgr := &rbac.Manager{}
	robotMgr := &robot2.Manager{}

	c := controller{robotMgr: robotMgr, rbacMgr: rbacMgr, proMgr: projectMgr}
	ctx := context.TODO()

	robotMgr.On("Update", mock.Anything, mock.Anything).Return(nil)
	projectMgr.On("Get", mock.Anything, mock.Anything).Return(&models.Project{ProjectID: 1, Name: "library"}, nil)
	rbacMgr.On("DeletePermissionByRole", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	rbacMgr.On("CreateRbacPolicy", mock.Anything, mock.Anything, mock.Anything).Return(int64(1), nil)
	rbacMgr.On("CreatePermission", mock.Anything, mock.Anything, mock.Anything).Return(int64(1), nil)

	err := c.Update(ctx, &Robot{
		Robot: model.Robot{
			Name:        "testcreate",
			Description: "testcreate",
			ExpiresAt:   0,
		},
		ProjectName: "library",
		Level:       LEVELPROJECT,
		Permissions: []*Permission{
			{
				Kind:      "project",
				Namespace: "library",
				Access: []*types.Policy{
					{
						Resource: "repository",
						Action:   "push",
					},
					{
						Resource: "repository",
						Action:   "pull",
					},
				},
			},
		},
	})
	suite.Nil(err)
}

func (suite *ControllerTestSuite) TestList() {
	projectMgr := &project.Manager{}
	rbacMgr := &rbac.Manager{}
	robotMgr := &robot2.Manager{}

	c := controller{robotMgr: robotMgr, rbacMgr: rbacMgr, proMgr: projectMgr}
	ctx := context.TODO()

	projectMgr.On("Get", mock.Anything, mock.Anything).Return(&models.Project{ProjectID: 1, Name: "library"}, nil)
	robotMgr.On("List", mock.Anything, mock.Anything).Return([]*model.Robot{
		{
			Name:        "test",
			Description: "test list method",
			ProjectID:   1,
			Secret:      utils.RandStringBytes(10),
		},
	}, nil)
	rbacMgr.On("GetPermissionsByRole", mock.Anything, mock.Anything, mock.Anything).Return([]*rbac_model.RolePermissions{
		{
			RoleType: ROBOTTYPE,
			RoleID:   1,
			Scope:    "/project/1",
			Resource: "repository",
			Action:   "pull",
		},
		{
			RoleType: ROBOTTYPE,
			RoleID:   1,
			Scope:    "/project/1",
			Resource: "repository",
			Action:   "push",
		},
	}, nil)
	projectMgr.On("Get", mock.Anything, mock.Anything).Return(&models.Project{ProjectID: 1, Name: "library"}, nil)
	rs, err := c.List(ctx, &q.Query{
		Keywords: map[string]interface{}{
			"name": "test3",
		},
	}, &Option{
		WithPermission: true,
	})
	suite.Nil(err)
	suite.Equal("project", rs[0].Permissions[0].Kind)
	suite.Equal("library", rs[0].Permissions[0].Namespace)
	suite.Equal("pull", rs[0].Permissions[0].Access[0].Action.String())
	suite.Equal("project", rs[0].Level)

}

func TestControllerTestSuite(t *testing.T) {
	suite.Run(t, &ControllerTestSuite{})
}
