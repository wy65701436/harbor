package security

import (
	"github.com/goharbor/harbor/src/common/security"
	robotCtx "github.com/goharbor/harbor/src/common/security/robot"
	"github.com/goharbor/harbor/src/common/utils"
	robot_ctl "github.com/goharbor/harbor/src/controller/robot"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/permission/types"

	"github.com/goharbor/harbor/src/pkg/robot/model"
	"net/http"
	"strings"
)

type robot2 struct{}

func (r *robot2) Generate(req *http.Request) security.Context {
	log := log.G(req.Context())
	name, secret, ok := req.BasicAuth()
	if !ok {
		return nil
	}
	if !strings.HasPrefix(name, config.RobotPrefix()) {
		return nil
	}
	key, err := config.SecretKey()
	if err != nil {
		log.Error("failed to get secret key")
		return nil
	}
	s, err := utils.ReversibleDecrypt(secret, key)
	if err != nil {
		return nil
	}
	log.Info("------------------")
	log.Info(s)
	log.Info("------------------")

	// TODO get the project name from the name patten
	robots, err := robot_ctl.Ctl.List(req.Context(), q.New(q.KeyWords{
		"name": strings.TrimPrefix(name, config.RobotPrefix()),
	}), &robot_ctl.Option{
		WithPermission: true,
	})
	if err != nil {
		return nil
	}
	if len(robots) == 0 {
		return nil
	}

	var accesses []*types.Policy
	robot := robots[0]
	if secret != robot.Secret {
		return nil
	}
	if robot.Disabled {
		return nil
	}
	// add the expiration check

	for _, p := range robot.Permissions {
		for _, a := range p.Access {
			access := &types.Policy{
				Resource: a.Resource,
				Action:   a.Action,
				Effect:   a.Effect,
			}
			accesses = append(accesses, access)
		}
	}

	modelRobot := &model.Robot{
		Name: robot.Name,
	}

	log.Debugf("a robot security context generated for request %s %s", req.Method, req.URL.Path)
	return robotCtx.NewSecurityContext(modelRobot, accesses)
}
