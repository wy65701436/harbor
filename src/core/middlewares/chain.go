package middlewares

import (
	"errors"
	"github.com/containous/alice"
	"github.com/goharbor/harbor/src/core/middlewares/contenttrust"
	"github.com/goharbor/harbor/src/core/middlewares/listrepo"
	"github.com/goharbor/harbor/src/core/middlewares/multiplmanifest"
	"github.com/goharbor/harbor/src/core/middlewares/readonly"
	"github.com/goharbor/harbor/src/core/middlewares/url"
	"github.com/goharbor/harbor/src/core/middlewares/vulnerable"
	"net/http"
)

type ChainBuilder struct {
	middlewares []string
}

func New(middlewares []string) *ChainBuilder {
	return &ChainBuilder{
		middlewares: middlewares,
	}
}

// CreateChain ...
func (b *ChainBuilder) CreateChain() *alice.Chain {
	chain := alice.New()
	for _, mName := range b.middlewares {
		middlewareName := mName
		chain = chain.Append(func(next http.Handler) (http.Handler, error) {
			constructor, err := b.getMiddleware(middlewareName)
			if err != nil {
				return nil, err
			}
			return constructor(next)
		})
	}
	return &chain
}

func (b *ChainBuilder) getMiddleware(mName string) (alice.Constructor, error) {
	var middleware alice.Constructor

	if mName == READONLY {
		middleware = func(next http.Handler) (http.Handler, error) {
			return readonly.New(next)
		}
	}
	if mName == URL {
		if middleware != nil {
			return nil, errors.New("middleware is not nil")
		}
		middleware = func(next http.Handler) (http.Handler, error) {
			return url.New(next)
		}
	}
	//if mName == REGQUOTA {
	//	if middleware != nil {
	//		return nil, errors.New("middleware is not nil")
	//	}
	//	middleware = func(next http.Handler) (http.Handler, error) {
	//		return regquota.New(next)
	//	}
	//}
	if mName == MUITIPLEMANIFEST {
		if middleware != nil {
			return nil, errors.New("middleware is not nil")
		}
		middleware = func(next http.Handler) (http.Handler, error) {
			return multiplmanifest.New(next)
		}
	}
	if mName == LISTREPO {
		if middleware != nil {
			return nil, errors.New("middleware is not nil")
		}
		middleware = func(next http.Handler) (http.Handler, error) {
			return listrepo.New(next)
		}
	}
	if mName == CONTENTTRUST {
		if middleware != nil {
			return nil, errors.New("middleware is not nil")
		}
		middleware = func(next http.Handler) (http.Handler, error) {
			return contenttrust.New(next)
		}
	}
	if mName == VULNERABLE {
		if middleware != nil {
			return nil, errors.New("middleware is not nil")
		}
		middleware = func(next http.Handler) (http.Handler, error) {
			return vulnerable.New(next)
		}
	}
	if middleware == nil {
		return nil, errors.New("no matched middleware")
	}

	return middleware, nil
}
