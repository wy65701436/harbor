package immutable

import (
	"github.com/docker/distribution"
	"github.com/goharbor/harbor/src/core/middlewares/util"
	"github.com/opencontainers/go-digest"

	"encoding/json"
	"fmt"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"math/rand"
	"net/http"
	"net/http/httptest"

	"github.com/goharbor/harbor/src/pkg/immutable/rule"
)

type HandlerSuite struct {
	suite.Suite
}

func doPutManifestRequest(projectID int64, projectName, name, tag, dgt string, next ...http.HandlerFunc) int {
	repository := fmt.Sprintf("%s/%s", projectName, name)

	url := fmt.Sprintf("/v2/%s/manifests/%s", repository, tag)
	req, _ := http.NewRequest("PUT", url, nil)

	mfInfo := &util.ManifestInfo{
		ProjectID:  projectID,
		Repository: repository,
		Tag:        tag,
		Digest:     dgt,
		References: []distribution.Descriptor{
			{Digest: digest.FromString(randomString(15))},
			{Digest: digest.FromString(randomString(15))},
		},
	}
	ctx := util.NewManifestInfoContext(req.Context(), mfInfo)
	rr := httptest.NewRecorder()

	var n http.HandlerFunc
	if len(next) > 0 {
		n = next[0]
	} else {
		n = func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusCreated)
		}
	}

	h := New(http.HandlerFunc(n))
	h.ServeHTTP(util.NewCustomResponseWriter(rr), req.WithContext(ctx))

	return rr.Code
}

func randomString(n int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyz"

	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}

	return string(b)
}

func (suite *HandlerSuite) addProject(projectName string) int64 {
	projectID, err := dao.AddProject(models.Project{
		Name:    projectName,
		OwnerID: 1,
	})
	suite.Nil(err, fmt.Sprintf("Add project failed for %s", projectName))
	return projectID
}

func (suite *HandlerSuite) addImmutableRule(pid int64) int64 {
	ir := &models.ImmutableRule{
		ProjectID: pid,
		Enabled:   true,
	}
	metadata := rule.Metadata{
		ID:       1,
		Priority: 1,
		Disabled: false,
		Action:   "immutable",
		Template: "immutable_template",
		TagSelectors: []*rule.Selector{
			{
				Kind:       "doublestar",
				Decoration: "matches",
				Pattern:    "release-[\\d\\.]+",
			},
		},
		ScopeSelectors: map[string][]*rule.Selector{
			"repository": {
				{
					Kind:       "doublestar",
					Decoration: "matches",
					Pattern:    "**",
				},
			},
		},
	}
	data, _ := json.Marshal(metadata)
	ir.TagFilter = string(data)

	id, err := dao.CreateImmutableRule(ir)
	require.NoError(suite.T(), err, "nil error expected but got %s", err)

	return id
}

func (suite *HandlerSuite) TestPutManifestCreated() {
	projectName := randomString(5)

	projectID := suite.addProject(projectName)
	immuRuleID := suite.addImmutableRule(projectID)
	defer func() {
		dao.DeleteProject(projectID)
		dao.DeleteImmutableRule(immuRuleID)
	}()

	dgt := digest.FromString(randomString(15)).String()
	code1 := doPutManifestRequest(projectID, projectName, "photon", "release-1.8.0", dgt)
	suite.Equal(http.StatusForbidden, code1)

	code2 := doPutManifestRequest(projectID, projectName, "photon", "latest", dgt)
	suite.Equal(http.StatusCreated, code2)

}
