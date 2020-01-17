package middleware

import (
	"github.com/goharbor/harbor/src/common/utils/test"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestParseManifestInfoFromPath(t *testing.T) {
	mustRequest := func(method, url string) *http.Request {
		req, _ := http.NewRequest(method, url, nil)
		return req
	}

	type args struct {
		req *http.Request
	}
	tests := []struct {
		name    string
		args    args
		want    *ManifestInfo
		wantErr bool
	}{
		{
			"ok for digest",
			args{mustRequest(http.MethodDelete, "/v2/library/photon/manifests/sha256:3e17b60ab9d92d953fb8ebefa25624c0d23fb95f78dde5572285d10158044059")},
			&ManifestInfo{
				ProjectID:  1,
				Repository: "library/photon",
				Digest:     "sha256:3e17b60ab9d92d953fb8ebefa25624c0d23fb95f78dde5572285d10158044059",
			},
			false,
		},
		{
			"ok for tag",
			args{mustRequest(http.MethodDelete, "/v2/library/photon/manifests/latest")},
			&ManifestInfo{
				ProjectID:  1,
				Repository: "library/photon",
				Tag:        "latest",
			},
			false,
		},
		{
			"project not found",
			args{mustRequest(http.MethodDelete, "/v2/notfound/photon/manifests/latest")},
			nil,
			true,
		},
		{
			"url not match",
			args{mustRequest(http.MethodDelete, "/v2/library/photon/manifest/latest")},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseManifestInfoFromPath(tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseManifestInfoFromPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseManifestInfoFromPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResolveManifest(t *testing.T) {
	test.InitDatabaseFromEnv()

	assert := assert.New(t)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})

	req := httptest.NewRequest(http.MethodDelete, "/v2/library/hello-world/manifests/latest", nil)
	rec := httptest.NewRecorder()
	ResolveManifest()(next).ServeHTTP(rec, req)
	assert.Equal(rec.Code, http.StatusOK)

	mf, ok := ManifestInfoFromContext(req.Context())
	assert.True(ok)
	assert.Equal(mf.Tag, "latest")
	assert.Equal(mf.ProjectID, 1)
}
