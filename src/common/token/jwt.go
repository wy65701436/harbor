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
	TokenID int64          `json:"ID"`
	Policy  []*rbac.Policy `json:"access"`
	jwt.StandardClaims
}

type HarborJWT struct {
	signMethod jwt.SigningMethod
	key        interface{}
	ttl        time.Duration
	issuer     string
}

// NewHarborJWT ...
func NewHarborJWT(optMap map[string]string) (*HarborJWT, error) {
	var opts JWTOptions
	err := opts.Parse(optMap)
	if err != nil {
		return nil, err
	}
	key, err := opts.GetKey()
	if err != nil {
		log.Errorf("failed to get key of token, %v", err)
	}
	htk := &HarborJWT{
		signMethod: opts.SignMethod,
		key:        key,
		ttl:        opts.TTL,
	}
	return htk, nil
}

// NewDefaultHarborJWT ...
func NewDefaultHarborJWT() (*HarborJWT, error) {
	var opts JWTOptions
	err := opts.Default()
	if err != nil {
		return nil, err
	}
	key, err := opts.GetKey()
	if err != nil {
		log.Errorf("failed to get key of token, %v", err)
		return nil, err
	}
	htk := &HarborJWT{
		signMethod: opts.SignMethod,
		key:        key,
		ttl:        opts.TTL,
		issuer:     opts.Issuer,
	}
	return htk, nil
}

// Encrypt ...
func (hj *HarborJWT) Encrypt(claims *Claims) (string, error) {
	claimsWrapper := Claims{
		claims.TokenID,
		claims.Policy,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(hj.ttl).Unix(),
			Issuer:    hj.issuer,
		},
	}
	tk := jwt.NewWithClaims(hj.signMethod, claimsWrapper)
	tokenStr, err := tk.SignedString(hj.key)
	if err != nil {
		log.Debugf(fmt.Sprintf("failed to issue token %v", err))
		return "", err
	}
	log.Infof(fmt.Sprintf("issued token: %s", tokenStr))
	return tokenStr, err
}

// Decrypt decrypt a string with key file to get the token claim info.
func (hj *HarborJWT) Decrypt(rawToken string) (*Claims, error) {
	claim := Claims{}
	token, err := jwt.ParseWithClaims(rawToken, &claim, func(token *jwt.Token) (interface{}, error) {
		if token.Method.Alg() != hj.signMethod.Alg() {
			return nil, errors.New("invalid signing method")
		}
		switch k := hj.key.(type) {
		case *rsa.PrivateKey:
			return &k.PublicKey, nil
		case *ecdsa.PrivateKey:
			return &k.PublicKey, nil
		default:
			return hj.key, nil
		}
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		fmt.Printf("%v %v", claims.Policy, claims.TokenID)
	} else {
		log.Errorf(fmt.Sprintf("parse token error, %v", err))
		return nil, err
	}
	return &claim, nil
}
