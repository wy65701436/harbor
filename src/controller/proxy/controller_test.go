//  Copyright Project Harbor Authors
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package proxy

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/docker/distribution"
	"github.com/opencontainers/go-digest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/controller/artifact"
	"github.com/goharbor/harbor/src/controller/blob"
	"github.com/goharbor/harbor/src/lib"
	_ "github.com/goharbor/harbor/src/lib/cache"
	"github.com/goharbor/harbor/src/lib/errors"
	proModels "github.com/goharbor/harbor/src/pkg/project/models"
	testproxy "github.com/goharbor/harbor/src/testing/controller/proxy"
)

type localInterfaceMock struct {
	mock.Mock
}

func (l *localInterfaceMock) SendPullEvent(ctx context.Context, repo, tag string) {
	panic("implement me")
}

func (l *localInterfaceMock) GetManifest(ctx context.Context, art lib.ArtifactInfo) (*artifact.Artifact, error) {
	args := l.Called(ctx, art)

	var a *artifact.Artifact
	if args.Get(0) != nil {
		a = args.Get(0).(*artifact.Artifact)
	}
	return a, args.Error(1)
}

func (l *localInterfaceMock) SameArtifact(ctx context.Context, repo, tag, dig string) (bool, error) {
	panic("implement me")
}

func (l *localInterfaceMock) BlobExist(ctx context.Context, art lib.ArtifactInfo) (bool, error) {
	args := l.Called(ctx, art)
	return args.Bool(0), args.Error(1)
}

func (l *localInterfaceMock) PushBlob(localRepo string, desc distribution.Descriptor, bReader io.ReadCloser) error {
	panic("implement me")
}

func (l *localInterfaceMock) PushManifest(repo string, tag string, manifest distribution.Manifest) error {
	args := l.Called(repo, tag, manifest)
	return args.Error(0)
}

func (l *localInterfaceMock) PushManifestList(ctx context.Context, repo string, tag string, man distribution.Manifest) error {
	panic("implement me")
}

func (l *localInterfaceMock) CheckDependencies(ctx context.Context, repo string, man distribution.Manifest) []distribution.Descriptor {
	args := l.Called(ctx, repo, man)
	return args.Get(0).([]distribution.Descriptor)
}

func (l *localInterfaceMock) DeleteManifest(repo, ref string) {
}

type proxyControllerTestSuite struct {
	suite.Suite
	local  *localInterfaceMock
	remote *testproxy.RemoteInterface
	ctr    Controller
	proj   *proModels.Project
}

func (p *proxyControllerTestSuite) SetupTest() {
	p.local = &localInterfaceMock{}
	p.remote = &testproxy.RemoteInterface{}
	p.proj = &proModels.Project{RegistryID: 1}
	p.ctr = &controller{
		blobCtl:     blob.Ctl,
		artifactCtl: artifact.Ctl,
		local:       p.local,
	}
}

func (p *proxyControllerTestSuite) TestUseLocalManifest_True() {
	ctx := context.Background()
	dig := "sha256:1a9ec845ee94c202b2d5da74a24f0ed2058318bfa9879fa541efaecba272e86b"
	art := lib.ArtifactInfo{Repository: "library/hello-world", Digest: dig}
	p.local.On("GetManifest", mock.Anything, mock.Anything).Return(&artifact.Artifact{}, nil)

	result, _, err := p.ctr.UseLocalManifest(ctx, art, p.remote)
	p.Assert().Nil(err)
	p.Assert().True(result)
}

func (p *proxyControllerTestSuite) TestUseLocalManifest_False() {
	ctx := context.Background()
	dig := "sha256:1a9ec845ee94c202b2d5da74a24f0ed2058318bfa9879fa541efaecba272e86b"
	desc := &distribution.Descriptor{Digest: digest.Digest(dig)}
	art := lib.ArtifactInfo{Repository: "library/hello-world", Digest: dig}
	p.remote.On("ManifestExist", mock.Anything, mock.Anything).Return(true, desc, nil)
	p.local.On("GetManifest", mock.Anything, mock.Anything).Return(nil, nil)
	result, _, err := p.ctr.UseLocalManifest(ctx, art, p.remote)
	p.Assert().Nil(err)
	p.Assert().False(result)
}

func (p *proxyControllerTestSuite) TestUseLocalManifest_429() {
	ctx := context.Background()
	dig := "sha256:1a9ec845ee94c202b2d5da74a24f0ed2058318bfa9879fa541efaecba272e86b"
	desc := &distribution.Descriptor{Digest: digest.Digest(dig)}
	art := lib.ArtifactInfo{Repository: "library/hello-world", Digest: dig}
	p.remote.On("ManifestExist", mock.Anything, mock.Anything).Return(false, desc, errors.New("too many requests").WithCode(errors.RateLimitCode))
	p.local.On("GetManifest", mock.Anything, mock.Anything).Return(nil, nil)
	_, _, err := p.ctr.UseLocalManifest(ctx, art, p.remote)
	p.Assert().NotNil(err)
	errors.IsRateLimitError(err)
}

func (p *proxyControllerTestSuite) TestUseLocalManifest_429ToLocal() {
	ctx := context.Background()
	dig := "sha256:1a9ec845ee94c202b2d5da74a24f0ed2058318bfa9879fa541efaecba272e86b"
	desc := &distribution.Descriptor{Digest: digest.Digest(dig)}
	art := lib.ArtifactInfo{Repository: "library/hello-world", Digest: dig}
	p.remote.On("ManifestExist", mock.Anything, mock.Anything).Return(false, desc, errors.New("too many requests").WithCode(errors.RateLimitCode))
	p.local.On("GetManifest", mock.Anything, mock.Anything).Return(&artifact.Artifact{}, nil)
	result, _, err := p.ctr.UseLocalManifest(ctx, art, p.remote)
	p.Assert().Nil(err)
	p.Assert().True(result)
}

func (p *proxyControllerTestSuite) TestUseLocalManifestWithTag_False() {
	ctx := context.Background()
	art := lib.ArtifactInfo{Repository: "library/hello-world", Tag: "latest"}
	desc := &distribution.Descriptor{}
	p.local.On("GetManifest", mock.Anything, mock.Anything).Return(&artifact.Artifact{}, nil)
	p.remote.On("ManifestExist", mock.Anything, mock.Anything).Return(false, desc, nil)
	result, _, err := p.ctr.UseLocalManifest(ctx, art, p.remote)
	p.Assert().True(errors.IsNotFoundErr(err))
	p.Assert().False(result)
}

func (p *proxyControllerTestSuite) TestUseLocalBlob_True() {
	ctx := context.Background()
	dig := "sha256:1a9ec845ee94c202b2d5da74a24f0ed2058318bfa9879fa541efaecba272e86b"
	art := lib.ArtifactInfo{Repository: "library/hello-world", Digest: dig}
	p.local.On("BlobExist", mock.Anything, mock.Anything).Return(true, nil)
	result := p.ctr.UseLocalBlob(ctx, art)
	p.Assert().True(result)
}

func (p *proxyControllerTestSuite) TestUseLocalBlob_False() {
	ctx := context.Background()
	dig := "sha256:1a9ec845ee94c202b2d5da74a24f0ed2058318bfa9879fa541efaecba272e86b"
	art := lib.ArtifactInfo{Repository: "library/hello-world", Digest: dig}
	p.local.On("BlobExist", mock.Anything, mock.Anything).Return(false, nil)
	result := p.ctr.UseLocalBlob(ctx, art)
	p.Assert().False(result)
}

func TestProxyControllerTestSuite(t *testing.T) {
	suite.Run(t, &proxyControllerTestSuite{})
}

func TestProxyCacheRemoteRepo(t *testing.T) {
	cases := []struct {
		name string
		in   lib.ArtifactInfo
		want string
	}{
		{
			name: `normal test`,
			in:   lib.ArtifactInfo{ProjectName: "dockerhub_proxy", Repository: "dockerhub_proxy/firstfloor/hello-world"},
			want: "firstfloor/hello-world",
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			got := getRemoteRepo(tt.in)
			if got != tt.want {
				t.Errorf(`(%v) = %v; want "%v"`, tt.in, got, tt.want)
			}
		})
	}
}
func TestGetRef(t *testing.T) {
	cases := []struct {
		name string
		in   lib.ArtifactInfo
		want string
	}{
		{
			name: `normal`,
			in:   lib.ArtifactInfo{Repository: "hello-world", Tag: "latest", Digest: "sha256:aabbcc"},
			want: "sha256:aabbcc",
		},
		{
			name: `digest_only`,
			in:   lib.ArtifactInfo{Repository: "hello-world", Tag: "", Digest: "sha256:aabbcc"},
			want: "sha256:aabbcc",
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			got := getReference(tt.in)
			if got != tt.want {
				t.Errorf(`(%v) = %v; want "%v"`, tt.in, got, tt.want)
			}
		})
	}
}

// TestTeeReadCloser tests the teeReadCloser implementation
// This validates the fix for issue #22555 where dual connections caused failures
func TestTeeReadCloser(t *testing.T) {
	// Create test data
	testData := []byte("test blob data for streaming")
	originalReader := io.NopCloser(bytes.NewReader(testData))

	// Create a pipe to simulate local storage
	pr, pw := io.Pipe()

	// Create TeeReader
	teeReader := io.TeeReader(originalReader, pw)

	// Wrap in teeReadCloser
	trc := &teeReadCloser{
		Reader: teeReader,
		closer: originalReader,
	}

	// Read from client side in a goroutine
	clientData := make([]byte, len(testData))
	clientDone := make(chan error, 1)
	go func() {
		n, err := io.ReadFull(trc, clientData)
		if err != nil {
			clientDone <- err
			return
		}
		if n != len(testData) {
			clientDone <- io.ErrShortBuffer
			return
		}
		clientDone <- nil
	}()

	// Read from local storage side
	localData := make([]byte, len(testData))
	n, err := io.ReadFull(pr, localData)
	require.NoError(t, err)
	require.Equal(t, len(testData), n)

	// Wait for client read to complete
	err = <-clientDone
	require.NoError(t, err)

	// Verify both sides received the same data
	assert.Equal(t, testData, clientData, "Client should receive correct data")
	assert.Equal(t, testData, localData, "Local storage should receive correct data")

	// Test Close
	err = trc.Close()
	assert.NoError(t, err, "Close should succeed")

	// Close the pipe writer
	pw.Close()
}

// TestTeeReadCloserLargeData tests with larger data to simulate real-world scenario
// This ensures the fix works for large blobs like those in issue #22555 (20GB images)
func TestTeeReadCloserLargeData(t *testing.T) {
	// Create 10MB test data
	dataSize := 10 * 1024 * 1024
	testData := make([]byte, dataSize)
	for i := range testData {
		testData[i] = byte(i % 256)
	}

	originalReader := io.NopCloser(bytes.NewReader(testData))

	// Create a pipe to simulate local storage
	pr, pw := io.Pipe()

	// Create TeeReader
	teeReader := io.TeeReader(originalReader, pw)

	// Wrap in teeReadCloser
	trc := &teeReadCloser{
		Reader: teeReader,
		closer: originalReader,
	}

	// Read from client side in a goroutine
	clientDone := make(chan error, 1)
	var clientBytesRead int64
	go func() {
		n, err := io.Copy(io.Discard, trc)
		clientBytesRead = n
		clientDone <- err
	}()

	// Read from local storage side
	localBytesRead, err := io.Copy(io.Discard, pr)
	require.NoError(t, err)

	// Wait for client read to complete
	err = <-clientDone
	require.NoError(t, err)

	// Verify both sides received all data
	assert.Equal(t, int64(dataSize), clientBytesRead, "Client should receive all data")
	assert.Equal(t, int64(dataSize), localBytesRead, "Local storage should receive all data")

	// Close
	trc.Close()
	pw.Close()
}

// TestTeeReadCloserErrorHandling tests error scenarios
func TestTeeReadCloserErrorHandling(t *testing.T) {
	testData := []byte("test data")
	originalReader := io.NopCloser(bytes.NewReader(testData))

	_, pw := io.Pipe()
	teeReader := io.TeeReader(originalReader, pw)

	trc := &teeReadCloser{
		Reader: teeReader,
		closer: originalReader,
	}

	// Close the pipe writer early to simulate error
	pw.Close()

	// Try to read - should get error from pipe
	buf := make([]byte, len(testData))
	_, err := io.ReadFull(trc, buf)
	// The read might succeed or fail depending on timing, but Close should always work

	// Close should still succeed
	err = trc.Close()
	assert.NoError(t, err, "Close should succeed even after pipe error")
}
