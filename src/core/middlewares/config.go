package middlewares

// const variables
const (
	READONLY         = "readonly"
	URL              = "url"
	MUITIPLEMANIFEST = "manifest"
	LISTREPO         = "listrepo"
	CONTENTTRUST     = "contenttrust"
	VULNERABLE       = "vulnerable"
	REGQUOTA         = "regquota"
)

// sequential organization
var Middlewares = []string{READONLY, URL, REGQUOTA, MUITIPLEMANIFEST, LISTREPO, CONTENTTRUST, VULNERABLE}
