package main

import (
	"fmt"
	helm_chart "helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"k8s.io/helm/pkg/chartutil"
	"net/http"
)

const (
	readmeFileName = "README.md"
	valuesFileName = "values.yaml"
)

type ChartVersionDetails struct {
	Metadata     *helm_repo.ChartVersion  `json:"metadata"`
	Dependencies []*helm_chart.Dependency `json:"dependencies"`
	Values       map[string]interface{}   `json:"values"`
	Files        map[string]string        `json:"files"`
}

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

		chartData, err := loader.LoadArchive(resp.Body)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		fmt.Println(chartData.Values)
		fmt.Println(chartData.Dependencies())
		fmt.Println(chartData.Files)
	}
	defer resp.Body.Close()

	return
}
