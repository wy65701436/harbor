package middlewares

import (
	"errors"
	"github.com/containous/alice"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/middlewares/contenttrust"
	"github.com/goharbor/harbor/src/core/middlewares/listrepo"
	"github.com/goharbor/harbor/src/core/middlewares/multiplmanifest"
	"github.com/goharbor/harbor/src/core/middlewares/readonly"
	"github.com/goharbor/harbor/src/core/middlewares/regquota"
	"github.com/goharbor/harbor/src/core/middlewares/url"
	"github.com/goharbor/harbor/src/core/middlewares/vulnerable"
	"net/http"
)

type DefaultCreator struct {
	middlewares []string
}

func New(middlewares []string) *DefaultCreator {
	return &DefaultCreator{
		middlewares: middlewares,
	}
}

// CreateChain ...
func (b *DefaultCreator) Create() *alice.Chain {
	chain := alice.New()
	for _, mName := range b.middlewares {
		middlewareName := mName
		chain = chain.Append(func(next http.Handler) http.Handler {
			constructor, err := b.getMiddleware(middlewareName)
			if err != nil {
				log.Error(err)
				return nil
			}
			return constructor(next)
		})
	}
	return &chain
}

func (b *DefaultCreator) getMiddleware(mName string) (alice.Constructor, error) {
	var middleware alice.Constructor

	if mName == READONLY {
		middleware = func(next http.Handler) http.Handler {
			return readonly.New(next)
		}
	}
	if mName == URL {
		if middleware != nil {
			return nil, errors.New("middleware is not nil")
		}
		middleware = func(next http.Handler) http.Handler {
			return url.New(next)
		}
	}
	if mName == REGQUOTA {
		if middleware != nil {
			return nil, errors.New("middleware is not nil")
		}
		middleware = func(next http.Handler) http.Handler {
			return regquota.New(next)
		}
	}
	if mName == MUITIPLEMANIFEST {
		if middleware != nil {
			return nil, errors.New("middleware is not nil")
		}
		middleware = func(next http.Handler) http.Handler {
			return multiplmanifest.New(next)
		}
	}
	if mName == LISTREPO {
		if middleware != nil {
			return nil, errors.New("middleware is not nil")
		}
		middleware = func(next http.Handler) http.Handler {
			return listrepo.New(next)
		}
	}
	if mName == CONTENTTRUST {
		if middleware != nil {
			return nil, errors.New("middleware is not nil")
		}
		middleware = func(next http.Handler) http.Handler {
			return contenttrust.New(next)
		}
	}
	if mName == VULNERABLE {
		if middleware != nil {
			return nil, errors.New("middleware is not nil")
		}
		middleware = func(next http.Handler) http.Handler {
			return vulnerable.New(next)
		}
	}
	if middleware == nil {
		return nil, errors.New("no matched middleware")
	}

	return middleware, nil
}
