package traceablehttp // import "github.com/Traceableai/goagent/instrumentation/net/traceablehttp"

import "github.com/hypertrace/goagent/sdk/instrumentation/net/http"

type HeaderAccessor interface {
	http.HeaderAccessor
}

func ShouldRecordBodyOfContentType(h HeaderAccessor) bool {
	return http.ShouldRecordBodyOfContentType(http.HeaderAccessor(h))
}
