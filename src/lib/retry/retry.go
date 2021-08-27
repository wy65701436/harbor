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

package retry

import (
<<<<<<< HEAD:src/lib/retry.go
	"errors"
	"fmt"
	"math"
=======
>>>>>>> d482a0c323781260a2f3a39145edc08de024ee9e:src/lib/retry/retry.go
	"math/rand"
	"time"

	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/jpillora/backoff"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

var (
	// ErrRetryTimeout timeout error for retrying
	ErrRetryTimeout = errors.New("retry timeout")
)

// Options options for the retry functions
type Options struct {
	InitialInterval time.Duration                        // the initial interval for retring after failure, default 100 milliseconds
	MaxInterval     time.Duration                        // the max interval for retring after failure, default 1 second
	Timeout         time.Duration                        // the total time before returning if something is wrong, default 1 minute
	Callback        func(err error, sleep time.Duration) // the callback function for Retry when the f called failed
}

// Option ...
type Option func(*Options)

// InitialInterval set initial interval
func InitialInterval(initial time.Duration) Option {
	return func(opts *Options) {
		opts.InitialInterval = initial
	}
}

// MaxInterval set max interval
func MaxInterval(max time.Duration) Option {
	return func(opts *Options) {
		opts.MaxInterval = max
	}
}

// Timeout set timeout interval
func Timeout(timeout time.Duration) Option {
	return func(opts *Options) {
		opts.Timeout = timeout
	}
}

// Callback set callback
func Callback(callback func(err error, sleep time.Duration)) Option {
	return func(opts *Options) {
		opts.Callback = callback
	}
}

// Retry retry until f run successfully or timeout
//
// NOTE: This function will use exponential backoff and jitter for retrying, see
// https://aws.amazon.com/blogs/architecture/exponential-backoff-and-jitter/ for more information
func Retry(f func() error, options ...Option) error {
	opts := &Options{}

	for _, o := range options {
		o(opts)
	}

	if opts.InitialInterval <= 0 {
		opts.InitialInterval = time.Millisecond * 100
	}

	if opts.MaxInterval <= 0 {
		opts.MaxInterval = time.Second
	}

	if opts.Timeout <= 0 {
		opts.Timeout = time.Minute
	}

	b := &backoff.Backoff{
		Min:    opts.InitialInterval,
		Max:    opts.MaxInterval,
		Factor: 2,
		Jitter: true,
	}

	var err error

	timeout := time.After(opts.Timeout)
	for {
		select {
		case <-timeout:
			return errors.New(ErrRetryTimeout).WithCause(err)
		default:
			err = f()
			if err == nil {
				return nil
			}

<<<<<<< HEAD:src/lib/retry.go
func getBackoff(attempt int, initialInterval, maxInterval time.Duration, equalJitter bool) time.Duration {
	max := float64(maxInterval)
	base := float64(initialInterval)

	dur := base * math.Pow(2, float64(attempt))
	fmt.Println("=========")
	fmt.Println(dur)
	fmt.Println(base)
	fmt.Println(attempt)
	fmt.Println("=========")
	if equalJitter {
		dur = dur/2 + float64(rand.Int63n(int64(dur))/2)
	}

	if dur < base {
		dur = base
	}
=======
			sleep := b.Duration()
			if opts.Callback != nil {
				opts.Callback(err, sleep)
			}
>>>>>>> d482a0c323781260a2f3a39145edc08de024ee9e:src/lib/retry/retry.go

			time.Sleep(sleep)
		}
	}
}
