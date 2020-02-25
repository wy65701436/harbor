package main

import (
	"fmt"
	helm_chart "helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"net/http"
)

const (
	readmeFileName = "README.md"
	valuesFileName = "values.yaml"
)

type ChartVersionDetails struct {
	Dependencies []*helm_chart.Dependency `json:"dependencies"`
	Values       map[string]interface{}   `json:"values"`
	Files        map[string]string        `json:"files"`
}

func main() {
	req, err := http.NewRequest("GET", "http://localhost:5000/v2/myrepo/mychart2/blobs/sha256:c145f3f6b48ca48aa04c32af3ba49bb2f6abbe410e9584b03c3e71bd2719428d", nil)
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

		//fmt.Println(chartData.Values)
		valueMap := make(map[string]interface{})
		readValue(chartData.Values, "", valueMap)
		fmt.Println(valueMap)
		//fmt.Println(chartData.Metadata.Dependencies)
		//fmt.Println(chartData.Files)
	}
	defer resp.Body.Close()

	return
}

// Recursively read value
func readValue(values map[string]interface{}, keyPrefix string, valueMap map[string]interface{}) {
	for key, value := range values {
		longKey := key
		if keyPrefix != "" {
			longKey = fmt.Sprintf("%s.%s", keyPrefix, key)
		}

		if subValues, ok := value.(map[string]interface{}); ok {
			readValue(subValues, longKey, valueMap)
		} else {
			valueMap[longKey] = value
		}
	}
}
