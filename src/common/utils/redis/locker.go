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
	"time"
)

const (
	lockCmd = `return redis.call('SET', KEYS[1], ARGV[1], 'NX', 'PX', ARGV[2])`
	unlockCmd = `if redis.call("get",KEYS[1]) == ARGV[1] then return redis.call("del",KEYS[1]) else return 0 end`
	dialConnectionTimeout = 30 * time.Second
	dialReadTimeout       = time.Minute + 10*time.Second
	dialWriteTimeout      = 10 * time.Second
)

type

func Lock(key, value string, timeoutMs int) (bool, error) {
	r := pool.Get()
	defer r.Close()

	cmd := redis.NewScript(1, lockCmd)
	if res, err := cmd.Do(r, key, value, timeoutMs); err != nil {
		return false, err
	} else {
		return res == "OK", nil
	}
}

// Unlock attempts to remove the lock on a key so long as the value matches.
// If the lock cannot be removed, either because the key has already expired or
// because the value was incorrect, an error will be returned.
func Unlock(key, value string) error {
	r := pool.Get()
	defer r.Close()

	cmd := redis.NewScript(1, unlockCmd}
	if res, err := redis.Int(cmd.Do(r, key, value)); err != nil {
		return err
	} else if res != 1 {
		return errors.New("Unlock failed, key or secret incorrect")
	}

	// Success
	return nil
}
