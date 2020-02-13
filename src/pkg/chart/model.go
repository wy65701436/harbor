package chartserver

import (
	"k8s.io/helm/pkg/chartutil"
	helm_repo "k8s.io/helm/pkg/repo"
	"time"
)

// ChartVersion extends the helm ChartVersion with additional labels
type ChartVersion struct {
	helm_repo.ChartVersion
}

// ChartVersions is an array of extended ChartVersion
type ChartVersions []*ChartVersion

// ChartVersionDetails keeps the detailed data info of the chart version
type ChartVersionDetails struct {
	Metadata     *helm_repo.ChartVersion `json:"metadata"`
	Dependencies []*chartutil.Dependency `json:"dependencies"`
	Values       map[string]interface{}  `json:"values"`
	Files        map[string]string       `json:"files"`
	Security     *SecurityReport         `json:"security"`
}

// SecurityReport keeps the info related with security
// e.g.: digital signature, vulnerability scanning etc.
type SecurityReport struct {
	Signature *DigitalSignature `json:"signature"`
}

// DigitalSignature used to indicate if the chart has been signed
type DigitalSignature struct {
	Signed     bool   `json:"signed"`
	Provenance string `json:"prov_file"`
}

// ChartInfo keeps the information of the chart
type ChartInfo struct {
	Name          string    `json:"name"`
	TotalVersions uint32    `json:"total_versions"`
	LatestVersion string    `json:"latest_version"`
	Created       time.Time `json:"created"`
	Updated       time.Time `json:"updated"`
	Icon          string    `json:"icon"`
	Home          string    `json:"home"`
	Deprecated    bool      `json:"deprecated"`
}
