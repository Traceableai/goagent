package propagation // import "github.com/Traceableai/goagent/propagation"

import (
	hyperpropagation "github.com/Traceableai/goagent/hypertrace/goagent/instrumentation/hypertrace/propagation"
)

type TextMapCarrier hyperpropagation.TextMapCarrier

var InjectTextMap = hyperpropagation.InjectTextMap

var ExtractTextMap = hyperpropagation.ExtractTextMap
