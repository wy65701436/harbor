package token

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/goharbor/harbor/src/common/utils/log"
	"time"
)

// HToken ...
type HToken struct {
	jwt.Token
	key interface{}
}

// NewWithClaims ...
func NewWithClaims(claims *RobotClaims) *HToken {
	dafaultOpts, err := NewDafaultOptions()
	if err != nil {
		log.Errorf(fmt.Sprintf("failed to get default jwt options %v", err))
		return nil
	}
	key, err := dafaultOpts.GetKey()
	if err != nil {
		return nil
	}
	rClaims := &RobotClaims{
		TokenID:   claims.TokenID,
		ProjectID: claims.ProjectID,
		Policy:    claims.Policy,
		StandardClaims: &jwt.StandardClaims{
			ExpiresAt: time.Now().Add(dafaultOpts.TTL).Unix(),
			Issuer:    dafaultOpts.Issuer,
		},
	}
	return &HToken{
		Token: *jwt.NewWithClaims(dafaultOpts.SignMethod, rClaims),
		key:   key,
	}
}

// SignedString get the SignedString.
func (htk *HToken) SignedString() (string, error) {
	raw, err := htk.Token.SignedString(htk.key)
	if err != nil {
		log.Debugf(fmt.Sprintf("failed to issue token %v", err))
		return "", err
	}
	log.Infof(fmt.Sprintf("issued token: %s", raw))
	return raw, err
}

// ParseWithClaims ...
func ParseWithClaims(rawToken string, claims jwt.Claims) (*HToken, error) {
	dafaultOpts, err := NewDafaultOptions()
	if err != nil {
		log.Errorf(fmt.Sprintf("failed to get default jwt options %v", err))
		return nil, err
	}
	key, err := dafaultOpts.GetKey()
	if err != nil {
		return nil, err
	}
	token, err := jwt.ParseWithClaims(rawToken, claims, func(token *jwt.Token) (interface{}, error) {
		if token.Method.Alg() != dafaultOpts.SignMethod.Alg() {
			return nil, errors.New("invalid signing method")
		}
		switch k := key.(type) {
		case *rsa.PrivateKey:
			return &k.PublicKey, nil
		case *ecdsa.PrivateKey:
			return &k.PublicKey, nil
		default:
			return key, nil
		}
	})
	if !token.Valid {
		log.Errorf(fmt.Sprintf("parse token error, %v", err))
		return nil, err
	}
	return &HToken{
		Token: *token,
		key:   key,
	}, nil
}
