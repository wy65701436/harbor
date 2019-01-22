package token

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNewDafaultOptions(t *testing.T) {
	defaultOpt, err := NewDafaultOptions()
	assert.NotNil(t, err)
	assert.Equal(t, defaultOpt.SignMethod, jwt.GetSigningMethod("RS256"))
	assert.Equal(t, defaultOpt.Issuer, "harbor-token-issuer")
	assert.Equal(t, defaultOpt.TTL, 60*time.Minute)
}

func TestGetKey(t *testing.T) {
	defaultOpt, err := NewDafaultOptions()
	assert.NotNil(t, err)
	key, err := defaultOpt.GetKey()
	assert.NotNil(t, err)
	assert.NotNil(t, key)
}
