package chartserver

import (
	"testing"

	htesting "github.com/goharbor/harbor/src/testing"
)

func TestGetChartDetails(t *testing.T) {
	chartOpr := ChartOperator{}
	chartDetails, err := chartOpr.GetChartDetails(htesting.HelmChartContent)
	if err != nil {
		t.Fatal(err)
	}

	if len(chartDetails.Dependencies) == 0 {
		t.Fatal("At least 1 dependency exitsing, but we got 0 now")
	}

	if len(chartDetails.Values) == 0 {
		t.Fatal("At least 1 value existing, but we got 0 now")
	}

	if chartDetails.Values["adminserver.adminPassword"] != "Harbor12345" {
		t.Fatalf("The value of 'adminserver.adminPassword' should be 'Harbor12345' but we got '%s' now", chartDetails.Values["adminserver.adminPassword"])
	}
}
