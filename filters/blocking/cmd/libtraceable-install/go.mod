module github.com/Traceableai/goagent/filters/blocking/cmd/libtraceable-install

go 1.16

replace github.com/Traceableai/goagent/filters/blocking/library => ../../library

require (
	github.com/Traceableai/goagent/filters/blocking/library v0.0.0-00010101000000-000000000000
	github.com/stretchr/testify v1.7.0
)
