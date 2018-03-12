package middlewares

import (
	"net/http"
)

// Handler ...
type Handler interface {
	ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc)
}

type inteceptor struct {
	handler Handler
	next *inteceptor
}

func (m inteceptor) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	m.handler.ServeHTTP(rw, r, m.next.ServeHTTP)
}

// HandlerFunc ...
type HandlerFunc func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) 

// ServeHTTP ...
func (h HandlerFunc) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	h(rw, r, next)
}

// Convert http handler to inteceptor
func Convert(handler http.Handler) Handler {
	return HandlerFunc(func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		handler.ServeHTTP(rw, r)
		next(rw, r)
	})
}

// InteceptorChain ...
type InteceptorChain struct {
	head inteceptor
	handlers []Handler
}

// NewInteceptorChain ...
func NewInteceptorChain(handlers ...Handler) *InteceptorChain {
	return &InteceptorChain {
		head: append(handlers),
		handlers: handlers,
	} 
}

func append(handlers []Handler) inteceptor {
	var next inteceptor
	if len(handlers) == 0 {
		return inteceptor{
			HandlerFunc(func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {}),
			&inteceptor{},
		}
	}else if len(handlers) > 1 {
		next = append(handlers[:1])
	}else{
		return inteceptor{
			HandlerFunc(func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {}),
			&inteceptor{},
		}		
	}

	return inteceptor{handlers[0], &next}
}

// Handlers ...
func (ic *InteceptorChain) Handlers() []Handler {
	ic.handlers
}
