package api

import (
	"errors"
	"net/http"

	"github.com/vmware/harbor/src/common/utils/log"
)

var ErrResgistryAbnormal = errors.New("registry is running abnormal.")

// Ping monitor the docker registry status
func Ping(w http.ResponseWriter, r *http.Request) {
	isRuning, err := IsRegRunning()
	if !isRuning {
		log.Errorf("Registry is not running : %v", err)
		handleInternalServerError(w)
		return
	}

	if err := writeJSON(w, "Pong"); err != nil {
		log.Errorf("Failed to write response: %v", err)
		return
	}
}

// IsRegRunning check the health status
func IsRegRunning() (bool, error) {
	addr := "http://127.0.0.1:5001/debug/health"
	resp, err := http.Get(addr)

	if err != nil {
		log.Errorf("Failed to validate registry status:%v : %v", addr, err)
		return false, err
	}
	if resp.StatusCode == 200 || resp.StatusCode == 401 {
		return true, nil
	}

	log.Errorf("Failed to validate registry status:%v : %v", addr, resp.StatusCode)
	return false, ErrResgistryAbnormal
}
