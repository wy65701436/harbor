package base

import (
	"github.com/goharbor/harbor/src/pkg/accessory/model"
	htesting "github.com/goharbor/harbor/src/testing"
	"github.com/stretchr/testify/suite"
	"testing"
)

type BaseTestSuite struct {
	htesting.Suite
	accessory model.Accessory
	digest    string
}

func (suite *BaseTestSuite) SetupSuite() {
	suite.digest = suite.DigestString()
	suite.accessory, _ = model.New(string(model.TypeNone), model.ArtifactID(1),
		model.SubjectArtifactID(2), model.Digest(suite.digest), model.Size(1234))
}

func (suite *BaseTestSuite) TestGetID() {
	suite.Equal(int64(0), suite.accessory.GetID())
}

func (suite *BaseTestSuite) TestGetArtID() {
	suite.Equal(int64(1), suite.accessory.GetArtifactID())
}

func (suite *BaseTestSuite) TestSubGetArtID() {
	suite.Equal(int64(2), suite.accessory.GetSubjectArtID())
}

func (suite *BaseTestSuite) TestSubGetSize() {
	suite.Equal(int64(1234), suite.accessory.GetSize())
}

func (suite *BaseTestSuite) TestSubGetDigest() {
	suite.Equal(suite.digest, suite.accessory.GetDigest())
}

func (suite *BaseTestSuite) TestSubGetType() {
	suite.Equal(model.TypeNone, suite.accessory.GetType())
}

func (suite *BaseTestSuite) TestSubGetRefType() {
	suite.Equal(model.RefNone, suite.accessory.Kind())
}

func (suite *BaseTestSuite) TestIsSoft() {
	suite.False(suite.accessory.IsSoft())
}

func (suite *BaseTestSuite) TestIsHard() {
	suite.False(suite.accessory.IsHard())
}

func TestCacheTestSuite(t *testing.T) {
	suite.Run(t, new(BaseTestSuite))
}
