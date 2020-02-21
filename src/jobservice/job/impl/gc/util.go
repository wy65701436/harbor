package gc

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
	"github.com/goharbor/harbor/src/common/utils/registry"
	"github.com/goharbor/harbor/src/common/utils/registry/auth"
	"net/http"
)

// delKeys ...
func delKeys(con redis.Conn, pattern string) error {
	iter := 0
	keys := make([]string, 0)
	for {
		arr, err := redis.Values(con.Do("SCAN", iter, "MATCH", pattern))
		if err != nil {
			return fmt.Errorf("error retrieving '%s' keys", pattern)
		}
		iter, err = redis.Int(arr[0], nil)
		if err != nil {
			return fmt.Errorf("unexpected type for Int, got type %T", err)
		}
		k, err := redis.Strings(arr[1], nil)
		if err != nil {
			return fmt.Errorf("converts an array command reply to a []string %v", err)
		}
		keys = append(keys, k...)

		if iter == 0 {
			break
		}
	}
	for _, key := range keys {
		_, err := con.Do("DEL", key)
		if err != nil {
			return fmt.Errorf("failed to clean registry cache %v", err)
		}
	}
	return nil
}

// deleteManifest calls the registry API to remove manifest
func deleteManifest(registryURL, repository, digest string) error {
	repoClient, err := newRepositoryClient(registryURL, repository)
	if err != nil {
		return err
	}
	_, exist, err := repoClient.ManifestExist(digest)
	if err != nil {
		return err
	}
	// it could be happened at remove manifest success but fail to delete harbor DB.
	// when the GC job executes again, the manifest should not exist.
	if !exist {
		return nil
	}
	if err := repoClient.DeleteManifest(digest); err != nil {
		return err
	}
	return nil
}

func newRepositoryClient(registryURL, repository string) (*registry.Repository, error) {
	uam := &auth.UserAgentModifier{
		UserAgent: "harbor-registry-client",
	}
	authorizer := auth.DefaultBasicAuthorizer()
	transport := registry.NewTransport(http.DefaultTransport, authorizer, uam)
	client := &http.Client{
		Transport: transport,
	}
	return registry.NewRepository(repository, registryURL, client)
}
