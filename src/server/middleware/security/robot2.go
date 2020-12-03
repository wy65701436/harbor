package security

import (
	"fmt"
	"github.com/goharbor/harbor/src/common/security"
	robotCtx "github.com/goharbor/harbor/src/common/security/robot"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/controller/project"
	robot_ctl "github.com/goharbor/harbor/src/controller/robot"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/permission/types"
	"regexp"
	"strings"
	"time"

	"github.com/goharbor/harbor/src/pkg/robot2/model"
	"net/http"
)

type robot2 struct{}

func (r *robot2) Generate(req *http.Request) security.Context {
	log := log.G(req.Context())
	var ok bool
	name, secret, ok := req.BasicAuth()
	if !ok {
		return nil
	}

	var projectName string
	var isProjectLevelRobot bool
	var robotName string
	robotName, ok = matchSysRobotPattern(name)
	if !ok {
		projectName, robotName, ok = matchProRobotPattern(name)
		if !ok {
			return nil
		}
		isProjectLevelRobot = true
	}

	query := &q.Query{}
	query.Keywords = q.KeyWords{
		"Name": robotName,
	}
	if isProjectLevelRobot {
		pro, err := project.Ctl.Get(req.Context(), projectName)
		if err != nil {
			log.Error(err)
			return nil
		}
		query.Keywords = q.KeyWords{
			"Name":      robotName,
			"ProjectID": pro.ProjectID,
		}
	}
	robots, err := robot_ctl.Ctl.List(req.Context(), query, &robot_ctl.Option{
		WithPermission: true,
	})
	if err != nil {
		log.Errorf("failed to list robots: %v", err)
		return nil
	}
	if len(robots) == 0 {
		log.Errorf("no robot found: %s", name)
		return nil
	}
	if len(robots) > 1 {
		log.Errorf("multiple robot entries found: %s", name)
		return nil
	}

	robot := robots[0]
	if utils.Encrypt(secret, robot.Salt, utils.SHA256) != robot.Secret {
		log.Errorf("failed to authenticate robot account: %s", name)
		return nil
	}
	if robot.Disabled {
		log.Errorf("failed to authenticate disabled robot account: %s", name)
		return nil
	}
	now := time.Now().Unix()
	if robot.ExpiresAt != -1 && robot.ExpiresAt <= now {
		log.Errorf("the robot account is expirated: %s", name)
		return nil
	}

	var accesses []*types.Policy
	for _, p := range robot.Permissions {
		for _, a := range p.Access {
			accesses = append(accesses, &types.Policy{
				Action:   a.Action,
				Effect:   a.Effect,
				Resource: types.Resource(fmt.Sprintf("%s/%s", p.Scope, a.Resource)),
			})
		}
	}

	modelRobot := &model.Robot{
		Name: robotName,
	}
	log.Infof("a robot2 security context generated for request %s %s", req.Method, req.URL.Path)
	return robotCtx.NewSecurityContext(modelRobot, robot.Level == robot_ctl.LEVELSYSTEM, accesses)
}

// matchSysRobotPattern match the name with system robot pattern and return the raw name
func matchSysRobotPattern(name string) (robot string, match bool) {
	name = strings.TrimPrefix(name, config.RobotPrefix())
	sysRobotNameReg := fmt.Sprintf("^(?P<robot>[a-z0-9]+(?:[._-][a-z0-9]+)*)$")
	strs := regexp.MustCompile(sysRobotNameReg).FindStringSubmatch(name)
	if len(strs) < 2 {
		return "", false
	}
	return strs[1], true
}

// matchProRobotPattern match the name with project robot pattern and return the project and raw robot name
func matchProRobotPattern(name string) (project, robot string, match bool) {
	name = strings.TrimPrefix(name, config.RobotPrefix())
	proRobotNameReg := fmt.Sprintf("^%s\\+%s", `(?P<project>[a-z0-9]+(?:[._-][a-z0-9]+)*)`, `(?P<robot>[a-z0-9]+(?:[._-][a-z0-9]+)*)$`)
	strs := regexp.MustCompile(proRobotNameReg).FindStringSubmatch(name)
	if len(strs) < 3 {
		return "", "", false
	}
	return strs[1], strs[2], true
}
