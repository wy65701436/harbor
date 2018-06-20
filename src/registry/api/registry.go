// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
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

package api

import (
	"bytes"
	"math/rand"
	"net/http"
	"os"
	"syscall"
	"time"

	"os/exec"

	"github.com/vmware/harbor/src/common/utils/log"
)

const (
	regConf    = "/etc/registry/config.yml"
	maxRetries = 5
)

// StartReg launch the docker registry.
func StartReg() error {
	var err error
	var isRunning bool

	isRunning, err = IsRegRunning()
	if isRunning {
		log.Info("docker registry is already running.")
		return nil
	}

	cmd := exec.Command("/bin/bash", "-c", "registry serve "+regConf)
	// Redirect the registry log to container stdout.
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := cmd.Start(); err != nil {
		log.Errorf("Fail to launch docker registry: %s", err)
		return err
	}

	time.Sleep(2 * time.Second)
	backoff := time.Second
	for i := 0; i < maxRetries; i++ {
		isRunning, err = IsRegRunning()
		if isRunning {
			monitorReg(cmd)
			log.Info("launch docker registry success")
			return nil
		}
		time.Sleep(backoff - time.Second + (time.Duration(rand.Int31n(1000)) * time.Millisecond))
		if i <= 4 {
			backoff = backoff * 2
		}
	}

	log.Errorf("Fail to launch docker registry: %s", err)
	return err
}

// monitorReg monitor the status of docker registry, quit container if it's crashed.
func monitorReg(c *exec.Cmd) {
	go func(c *exec.Cmd) {
		if err := c.Wait(); err != nil {
			log.Warningf("Docker regsitry is crashed, err: %s, quit the container to restart it.", err)
			os.Exit(1)
		}
	}(c)
}

// StartGC ...
func StartGC(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	cmd := exec.Command("/bin/bash", "-c", "registry garbage-collect "+regConf)
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	if err := cmd.Run(); err != nil {
		log.Errorf("Fail to execute GC: %s", errBuf.String())
		gcr := GCResult{false, errBuf.String(), start, time.Now()}
		gcr.DumpGCResult()
		handleInternalServerError(w)
	}

	gcr := GCResult{true, outBuf.String(), start, time.Now()}
	gcr.DumpGCResult()
	if err := writeJSON(w, gcr); err != nil {
		log.Errorf("failed to write response: %v", err)
		return
	}
}
