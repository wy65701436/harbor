package token

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/config"
	"io/ioutil"
	"time"
)

var (
	defaultTTL          = 60 * time.Minute
	defaultIssuer       = "harbor-token-issuer"
	defaultSignedMethod = "RS256"
	defaultPrivateKey   = config.TokenPrivateKeyPath()
	//defaultPrivateKey = "/Users/yan/go/src/github.com/goharbor/harbor/make/common/config/core/private_key.pem"
)

type JWTOptions struct {
	SignMethod jwt.SigningMethod
	PublicKey  []byte
	PrivateKey []byte
	TTL        time.Duration
	Issuer     string
}

func (jop *JWTOptions) Parse(opts map[string]string) error {
	var err error
	if opts["ttl"] == "" {
		jop.TTL = defaultTTL
	}
	if opts["issuer"] == "" {
		jop.Issuer = defaultIssuer
	}

	signedMethod := opts["signedmethod"]
	if signedMethod == "" {
		jop.SignMethod = jwt.GetSigningMethod(defaultSignedMethod)
	} else {
		jop.SignMethod = jwt.GetSigningMethod(signedMethod)
	}
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
			fmt.Printf("failed to read public key %v", err)
			log.Errorf(fmt.Sprintf("failed to read public key %v", err))
			return err
		}
	}
	return nil
}

func (jop *JWTOptions) Default() error {
	var err error
	jop.TTL = defaultTTL
	jop.Issuer = defaultIssuer
	jop.SignMethod = jwt.GetSigningMethod(defaultSignedMethod)
	jop.PrivateKey, err = ioutil.ReadFile(defaultPrivateKey)
	if err != nil {
		log.Errorf(fmt.Sprintf("failed to read private key %v", err))
		return err
	}
	return nil
}

func (jop *JWTOptions) GetKey() (interface{}, error) {
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
		if privateKey == nil {
			if publicKey == nil {
				// Neither key given
				return nil, fmt.Errorf("key is missing.")
			}
			// Public key only, can verify tokens
			return publicKey, nil
		}
		if publicKey != nil && publicKey.E != privateKey.E && publicKey.N.Cmp(privateKey.N) != 0 {
			return nil, fmt.Errorf("the public key and private key are not match.")
		}
		return privateKey, nil
	default:
		return nil, fmt.Errorf(fmt.Sprintf("unsupported sign method, %s", jop.SignMethod))
	}
	return nil, nil
}
