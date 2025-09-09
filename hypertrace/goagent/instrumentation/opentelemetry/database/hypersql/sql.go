package hypersql // import "github.com/Traceableai/goagent/hypertrace/goagent/instrumentation/opentelemetry/database/hypersql"

import (
	"database/sql/driver"

	"github.com/Traceableai/goagent/hypertrace/goagent/instrumentation/opentelemetry"
	sdkSQL "github.com/Traceableai/goagent/hypertrace/goagent/sdk/instrumentation/database/sql"
)

// Wrap takes a SQL driver and wraps it with Hypertrace instrumentation.
func Wrap(d driver.Driver, options *sdkSQL.Options) driver.Driver {
	return sdkSQL.Wrap(d, opentelemetry.StartSpan, options)
}

// Register initializes and registers the hypersql wrapped database driver
// identified by its driverName. On success it
// returns the generated driverName to use when calling hypersql.Open.
func Register(driverName string, options *sdkSQL.Options) (string, error) {
	return sdkSQL.Register(driverName, opentelemetry.StartSpan, options)
}
