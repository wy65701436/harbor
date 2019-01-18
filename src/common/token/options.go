package token

import (
	"time"
	"github.com/dgrijalva/jwt-go"
	"errors"
	"io/ioutil"
	"github.com/goharbor/harbor/src/common/utils/log"
	"fmt"
	"crypto/rsa"
)

var (
	// DefaultTTL is 5 minutes for testing
	DefaultTTL = 5 * time.Minute
)

type jwtOptions struct {
	SignMethod jwt.SigningMethod
	PublicKey  []byte
	PrivateKey []byte
	TTL        time.Duration
}

func (jop *jwtOptions) Parse(opts map[string]string) error {
	var err error
	if opts["ttl"] == "" {
		jop.TTL = DefaultTTL
	}

	signedMethod := opts["signedmethod"]
	jop.SignMethod = jwt.GetSigningMethod(signedMethod)
	if jop.SignMethod == nil {
		return errors.New("Not valid signed method")
	}

	privateKeyFile := opts["privatekey"]
	if privateKeyFile != "" {
		jop.PrivateKey, err = ioutil.ReadFile(privateKeyFile)
		if err != nil {
			log.Errorf(fmt.Sprintf("failed to read private key %v", err))
			return err
		}
	}

	publicKeyFile := opts["publickey"]
	if publicKeyFile != "" {
		jop.PublicKey, err = ioutil.ReadFile(publicKeyFile)
		if err != nil {
			log.Errorf(fmt.Sprintf("failed to read public key %v", err))
			return err
		}
	}
	return nil
}

func (jop *jwtOptions) GetKey() (interface{}, error) {
	var err error
	var privateKey *rsa.PrivateKey
	var publicKey *rsa.PublicKey

	switch jop.SignMethod.(type) {
	case *jwt.SigningMethodRSA, *jwt.SigningMethodRSAPSS:
		if len(jop.PrivateKey) > 0 {
			privateKey, err = jwt.ParseRSAPrivateKeyFromPEM(jop.PrivateKey)
			if err != nil {
				return nil, err
			}
		}

		if len(jop.PublicKey) > 0 {
			publicKey, err = jwt.ParseRSAPublicKeyFromPEM(jop.PublicKey)
			if err != nil {
				return nil, err
			}
		}

		if privateKey != nil {
			if publicKey != nil {
				if publicKey.E != privateKey.E && publicKey.N.Cmp(privateKey.N) != 0 {
					return nil, fmt.Errorf("the public key and private key are not match.")
				}
			}
			return privateKey, nil
		}

		return nil, fmt.Errorf("no key provided.")

	default:
		return nil, fmt.Errorf(fmt.Sprintf("unsupported sign method, %s", jop.SignMethod))
	}
	return nil, nil
}
