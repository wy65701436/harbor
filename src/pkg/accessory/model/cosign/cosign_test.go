package cosign

import (
	"github.com/goharbor/harbor/src/pkg/accessory/model"
	"github.com/goharbor/harbor/src/pkg/accessory/model/base"
	htesting "github.com/goharbor/harbor/src/testing"
	"github.com/stretchr/testify/suite"
	"testing"
)

type CosignTestSuite struct {
	htesting.Suite
	accessory model.Accessory
	digest    string
}

func (suite *CosignTestSuite) SetupSuite() {
	suite.digest = suite.DigestString()
	suite.accessory = &Signature{base.Default{
		Opts: &model.Options{
			ArtifactID:        1,
			SubjectArtifactID: 2,
			Size:              1234,
			Digest:            suite.digest,
		},
	}}
}

func (suite *CosignTestSuite) TestGetID() {
	suite.Equal(int64(0), suite.accessory.GetID())
}

func (suite *CosignTestSuite) TestGetArtID() {
	suite.Equal(int64(1), suite.accessory.GetArtifactID())
}

func (suite *CosignTestSuite) TestSubGetArtID() {
	suite.Equal(int64(2), suite.accessory.GetSubjectArtID())
}

func (suite *CosignTestSuite) TestSubGetSize() {
	suite.Equal(int64(1234), suite.accessory.GetSize())
}

func (suite *CosignTestSuite) TestSubGetDigest() {
	suite.Equal(suite.digest, suite.accessory.GetDigest())
}

func (suite *CosignTestSuite) TestSubGetType() {
	suite.Equal(model.TypeCosignSignature, suite.accessory.GetType())
}

func (suite *CosignTestSuite) TestSubGetRefType() {
	suite.Equal(model.RefHard, suite.accessory.Kind())
}

func (suite *CosignTestSuite) TestIsSoft() {
	suite.False(suite.accessory.IsSoft())
}

func (suite *CosignTestSuite) TestIsHard() {
	suite.True(suite.accessory.IsHard())
}

func TestCacheTestSuite(t *testing.T) {
	suite.Run(t, new(CosignTestSuite))
}
