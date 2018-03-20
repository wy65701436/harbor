package middlewares

import "net/http"

var middlewares []http.Handler

func Handle(rw http.ResponseWriter, req *http.Request) {
	middlewares = []http.Handler{
		&ReadonlyHandler{},
		&UrlHandler{},
		&ListReposHandler{}}

	for _, h := range middlewares {
		h.ServeHTTP(rw, req)
	}
}
