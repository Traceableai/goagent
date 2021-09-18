package goagent // import "github.com/Traceableai/goagent"

import (
	"context"
	"time"

	"github.com/hypertrace/goagent/instrumentation/hypertrace"
	"github.com/hypertrace/goagent/sdk"
)

type Span interface {
	sdk.Span
}

func SpanFromContext(ctx context.Context) Span {
	return hypertrace.SpanFromContext(ctx)
}

type Option func(o *sdk.SpanOptions)

type SpanKind string

const (
	Undetermined SpanKind = SpanKind(sdk.Undetermined)
	Client       SpanKind = SpanKind(sdk.Client)
	Server       SpanKind = SpanKind(sdk.Server)
	Producer     SpanKind = SpanKind(sdk.Producer)
	Consumer     SpanKind = SpanKind(sdk.Consumer)
)

func WithSpanKind(kind SpanKind) Option {
	return func(o *sdk.SpanOptions) {
		o.Kind = sdk.SpanKind(kind)
	}
}

func WithTimestamp(ts time.Time) Option {
	return func(o *sdk.SpanOptions) {
		o.Timestamp = ts
	}
}

func htStarterToSpanStarter(
	s func(ctx context.Context, name string, opts *sdk.SpanOptions) (context.Context, sdk.Span, func()),
) func(ctx context.Context, name string, opts ...Option) (context.Context, Span, func()) {
	return func(ctx context.Context, name string, opts ...Option) (context.Context, Span, func()) {
		o := &sdk.SpanOptions{}
		for _, opt := range opts {
			opt(o)
		}

		return s(ctx, name, o)
	}
}

var StartSpan = htStarterToSpanStarter(hypertrace.StartSpan)
