package debug

import (
	"fmt"
	"helm.sh/helm/v3/pkg/chart/loader"
	"io/ioutil"
	"net/http"
)

func main() {
	req, err := http.NewRequest("GET", "http://localhost:5000/v2/myrepo/mychart/blobs/sha256:0bd64cfb958b68c71b46597e22185a41e784dc96e04090bc7d2a480b704c3b65", nil)
	if err != nil {
		return
	}

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Sprint(err.Error())
		return
	}

	if resp.StatusCode == http.StatusOK {
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return
		}
		fmt.Sprint(string(b))
	}
	defer resp.Body.Close()

	return
}
