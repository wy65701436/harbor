package token

import (
	"github.com/dgrijalva/jwt-go"
	"time"
	"github.com/goharbor/harbor/src/common/utils/log"
	"fmt"
)

var keyFile = ""

type HarborClaims struct {
	jwt.StandardClaims
	Access  *[]Policy `json:"access"`
	TokenID int64
}

type HarborPermClaim struct {
	TokenID int64
	Access  *[]Policy `json:"access"`
}

func (hpc *HarborPermClaim) Valid() error {
	return nil
}

// Resource the type of resource
type Resource string

func (res Resource) String() string {
	return string(res)
}

// Action the type of action
type Action string

func (act Action) String() string {
	return string(act)
}

type Policy struct {
	Resource
	Action
}

type HarborToken struct {
	signMethod jwt.SigningMethod
	key        interface{}
	ttl        time.Duration
}

func (ht *HarborToken) Issue(policy []Policy) (string, error) {

	// Future work: let a jwt token include permission information would be useful for
	// permission checking in proxy side.
	tk := jwt.NewWithClaims(ht.signMethod,
		&HarborClaims{
			Access:  &policy,
			TokenID: 1,
		})

	token, err := tk.SignedString(ht.key)
	if err != nil {
		log.Debug(fmt.Print("failed to issue token %v", err))
		return "", err
	}

	log.Infof(fmt.Sprintf("jwt token: %s", token))
	return token, err
}
