// Copyright 2018 Project Harbor Authors
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

package redis

import (
	"fmt"
	"time"

	"github.com/garyburd/redigo/redis"
	"gopkg.in/redsync.v1"
)

func newPool(server string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,

		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", server)
			if err != nil {
				return nil, err
			}
			return c, err
		},

		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}

func newPools(servers []string) []redsync.Pool {
	pools := []redsync.Pool{}
	for _, server := range servers {
		pool := newPool(server)
		pools = append(pools, pool)
	}

	return pools
}

func main() {
	pools := newPools([]string{"127.0.0.1:6379", "127.0.0.1:6378", "127.0.0.1:6377"})
	rs := redsync.New(pools)
	m := rs.NewMutex("/lock")

	err := m.Lock()
	if err != nil {
		panic(err)
	}
	fmt.Println("lock success")
	unlockRes := m.Unlock()
	fmt.Println("unlock result: ", unlockRes)

}
