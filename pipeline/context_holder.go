package pipeline

import (
	"golang.org/x/net/context"
)

type ContextHolder interface {
	Context() context.Context
}
