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

package metadata

import (
	"context"
	"time"

	"github.com/goharbor/harbor/src/common/security"
	event2 "github.com/goharbor/harbor/src/controller/event"
	"github.com/goharbor/harbor/src/controller/robot"
	"github.com/goharbor/harbor/src/pkg/notifier/event"
)

// CreateRobotEventMetadata is the metadata from which the create robot event can be resolved
type CreateRobotEventMetadata struct {
	Ctx   context.Context
	Robot *robot.Robot
}

// Resolve to the event from the metadata
func (c *CreateRobotEventMetadata) Resolve(event *event.Event) error {
	data := &event2.CreateRobotEvent{
		EventType: event2.TopicCreateRobot,
		Robot:     c.Robot,
		OccurAt:   time.Now(),
	}
	cx, exist := security.FromContext(c.Ctx)
	if exist {
		data.Operator = cx.GetUsername()
	}
	event.Topic = event2.TopicCreateRobot
	event.Data = data
	return nil
}
