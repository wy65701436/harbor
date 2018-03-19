package middlewares

type contextKey string

const imageInfoCtxKey = contextKey("ImageInfo")

type ImageInfo struct {
	repository  string
	reference   string
	projectName string
	digest      string
}
