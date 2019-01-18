package token

import (
	"github.com/dgrijalva/jwt-go"
	"time"
	"github.com/goharbor/harbor/src/common/utils/log"
	"fmt"
	"github.com/goharbor/harbor/src/common/rbac"
	"crypto/rsa"
	"crypto/ecdsa"
	"errors"
)

type Claims struct {
	jwt.StandardClaims
	TokenID int64          `json:"ID"`
	Policy  *[]rbac.Policy `json:"access"`
}

type HarborToken struct {
	signMethod jwt.SigningMethod
	key        interface{}
	ttl        time.Duration
}

// NewHarborToken ...
func NewHarborToken(optMap map[string]string) (*HarborToken, error) {
	var opts jwtOptions
	var htk *HarborToken
	err := opts.Parse(optMap)
	if err != nil {
		return nil, err
	}
	key, err := opts.GetKey()
	if err != nil {
		log.Errorf("failed to get key of token, %v", err)
	}

	htk = &HarborToken{
		signMethod: opts.SignMethod,
		key:        key,
		ttl:        opts.TTL,
	}

	return htk, nil
}

func (htk *HarborToken) Encrypt(claims Claims) (string, error) {
	tk := jwt.NewWithClaims(htk.signMethod, claims)
	tokenStr, err := tk.SignedString(htk.key)
	if err != nil {
		log.Debugf(fmt.Sprintf("failed to issue token %v", err))
		return "", err
	}
	log.Infof(fmt.Sprintf("issued token: %s", tokenStr))
	return tokenStr, err
}

func (htk *HarborToken) Decrypt(rawToken string) (*Claims, error) {
	var claim Claims
	token, err := jwt.ParseWithClaims(rawToken, claim, func(token *jwt.Token) (interface{}, error) {
		if token.Method.Alg() != htk.signMethod.Alg() {
			return nil, errors.New("invalid signing method")
		}
		switch k := htk.key.(type) {
		case *rsa.PrivateKey:
			return &k.PublicKey, nil
		case *ecdsa.PrivateKey:
			return &k.PublicKey, nil
		default:
			return htk.key, nil
		}
	})

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		fmt.Printf("%v %v", claims.Policy, claims.StandardClaims.ExpiresAt)
	} else {
		log.Errorf(fmt.Sprintf("parse token error, %v", err))
		fmt.Println(err)
		return nil, err
	}
	return &claim, nil
}

func main() {
	fmt.Println("test token string ...")
	return
}
git 