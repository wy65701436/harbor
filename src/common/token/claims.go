package token

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/pkg/errors"
)

// RobotClaims RobotClaims implements the interface of jwt.Claims
type RobotClaims struct {
	TokenID        int64          `json:"ID"`
	ProjectID      int64          `json:"PID"`
	Policy         []*rbac.Policy `json:"Access"`
	StandardClaims *jwt.StandardClaims
}

// Validates time based claims "id, product and access".
func (rc RobotClaims) Valid() error {
	if rc.TokenID < 0 {
		return errors.New("Token id must an valid INT.")
	}
	if rc.ProjectID < 0 {
		return errors.New("Token id must an valid INT.")
	}
	if rc.Policy == nil {
		return errors.New("The access info cannot be nil.")
	}
	return nil
}
