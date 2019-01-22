package main

import (
	"github.com/goharbor/harbor/src/common/rbac"
	"fmt"
)

func main() {
	rbacPolicy := &rbac.Policy{
		Resource: "/project/libray/repository",
		Action:   "pull",
	}
	policies := []*rbac.Policy{}
	policies = append(policies, rbacPolicy)

	policy := &RobotClaims{
		TokenID:  1,
		PublicID: 1,
		Policy:   policies,
	}

	token := NewWithClaims(policy)

	rawTk, err := token.SignedString()
	if err != nil {
		fmt.Sprintf("get error on signingn, %v", err)
	}
	fmt.Sprintf("get raw token, %s", rawTk)
}
