package security

import (
	"github.com/goharbor/harbor/src/common"
	core_cfg "github.com/goharbor/harbor/src/core/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

func TestRobot2(t *testing.T) {
	conf := map[string]interface{}{
		common.RobotNamePrefix: "robot@",
	}
	core_cfg.InitWithSettings(conf)

	robot := &robot2{}
	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1/api/projects/", nil)
	require.Nil(t, err)
	req.SetBasicAuth("robot@est1", "Harbor12345")
	ctx := robot.Generate(req)
	assert.Nil(t, ctx)
}

func TestMatchSysRobotPattern(t *testing.T) {
	conf := map[string]interface{}{
		common.RobotNamePrefix: "robot$",
	}
	core_cfg.InitWithSettings(conf)

	r, ok := matchSysRobotPattern("robot$test1")
	assert.True(t, ok)
	assert.Equal(t, "test1", r)

	r, ok = matchSysRobotPattern("robot$q1w2e3r4")
	assert.True(t, ok)
	assert.Equal(t, "q1w2e3r4", r)

	r, ok = matchSysRobotPattern("robot@test1")
	assert.False(t, ok)

	conf = map[string]interface{}{
		common.RobotNamePrefix: "mysettting&",
	}
	core_cfg.InitWithSettings(conf)

	r, ok = matchSysRobotPattern("robot$q1w2e3r4")
	assert.False(t, ok)

	r, ok = matchSysRobotPattern("mysettting&test122")
	assert.True(t, ok)
	assert.Equal(t, "test122", r)
}

func TestMatchProRobotPattern(t *testing.T) {
	conf := map[string]interface{}{
		common.RobotNamePrefix: "robot$",
	}
	core_cfg.InitWithSettings(conf)

	p, r, ok := matchProRobotPattern("robot$library+test1")
	assert.True(t, ok)
	assert.Equal(t, "library", p)
	assert.Equal(t, "test1", r)

	p, r, ok = matchProRobotPattern("robot$harbor.test+robot.test")
	assert.True(t, ok)
	assert.Equal(t, "harbor.test", p)
	assert.Equal(t, "robot.test", r)

	p, r, ok = matchProRobotPattern("robot$harbor_test+robot_test.account")
	assert.True(t, ok)
	assert.Equal(t, "harbor_test", p)
	assert.Equal(t, "robot_test.account", r)

	_, _, ok = matchProRobotPattern("robot@library+test1")
	assert.False(t, ok)

	conf = map[string]interface{}{
		common.RobotNamePrefix: "mysetting!@#$",
	}
	core_cfg.InitWithSettings(conf)

	_, _, ok = matchProRobotPattern("robot@library+test1")
	assert.False(t, ok)

	p, r, ok = matchProRobotPattern("mysetting!@#$harbor_test+robot_test.account")
	assert.True(t, ok)
	assert.Equal(t, "harbor_test", p)
	assert.Equal(t, "robot_test.account", r)
}
