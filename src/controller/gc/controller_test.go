package gc

import (
	schedulertesting "github.com/goharbor/harbor/src/testing/pkg/scheduler"
	tasktesting "github.com/goharbor/harbor/src/testing/pkg/task"
	"github.com/stretchr/testify/suite"
)

type gcCtrTestSuite struct {
	suite.Suite
	scheduler *schedulertesting.Scheduler
	execMgr   *tasktesting.FakeExecutionManager
	taskMgr   *tasktesting.FakeManager
}

func (s *schedulerTestSuite) SetupTest() {
	registry = map[string]CallbackFunc{}
	err := RegisterCallbackFunc("callback", func(interface{}) error { return nil })
	s.Require().Nil(err)

	s.dao = &mockDAO{}
	s.execMgr = &tasktesting.FakeExecutionManager{}
	s.taskMgr = &tasktesting.FakeManager{}w

	s.scheduler = &scheduler{
		dao:     s.dao,
		execMgr: s.execMgr,
		taskMgr: s.taskMgr,
	}
}
