package traceablehttp // import "github.com/Traceableai/goagent/instrumentation/net/traceablehttp"

import (
	"github.com/hypertrace/goagent/sdk/filter"
	"github.com/hypertrace/goagent/sdk/instrumentation/net/http"
)

type Options struct {
	Filter filter.Filter
}

func (o *Options) toSDKOptions() *http.Options {
	if o == nil {
		return &http.Options{}
	}

	return &http.Options{
		Filter: o.Filter,
	}
}
