module github.com/traceableai/goagent

go 1.16

require (
	github.com/hypertrace/agent-config/gen/go v0.0.0-20210827180927-f8a7187ff6cc
	github.com/hypertrace/goagent v0.3.1-0.20210827201008-0ff22ae72e11
	github.com/stretchr/testify v1.7.0
	github.com/traceableai/agent-config/gen/go v0.0.0-00010101000000-000000000000
	google.golang.org/protobuf v1.27.1
)

replace github.com/traceableai/agent-config/gen/go => ./config/_traceableai/agent-config/ai/traceable/agent/config
