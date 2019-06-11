package middlewares

import "github.com/containous/alice"

type ChainCreator interface {
	Create(middlewares []string) *alice.Chain
}
