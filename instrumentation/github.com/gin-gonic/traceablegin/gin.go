package traceablegin // import "github.com/Traceableai/goagent/instrumentation/github.com/gin-gonic/traceablegin"

import (
	"github.com/Traceableai/goagent/instrumentation/internal/traceablefilter"
	"github.com/gin-gonic/gin"
	"github.com/hypertrace/goagent/instrumentation/hypertrace/github.com/gin-gonic/hypergin"
)

func Middleware(opts ...Option) gin.HandlerFunc {
	o := &options{}
	for _, opt := range opts {
		opt(o)
	}
	o.Filter = traceablefilter.AppendTraceableFilter(o.Filter)

	return hypergin.Middleware(o.translateOptions()...)
}
