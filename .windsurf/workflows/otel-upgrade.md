---
description: Upgrade OpenTelemetry Go + go-contrib dependencies
---

# Goal
Upgrade `go.opentelemetry.io/otel` and related `go.opentelemetry.io/contrib` modules in the root `go.mod` file in this repo, and apply any required Traceable-specific code updates (notably around the custom `batchspanprocessor`). After, `make tidy` will update the other `go.mod` files in the examples, `_tests` and tools directories.

# Prereqs
- `go` installed.
- The repos `github.com/open-telemetry/opentelemetry-go` and `github.com/open-telemetry/opentelemetry-go-contrib` should be cloned in the same directory as this repo or should be somewhere in a subdirectory of the parent directory of this repo. If you cannot find them, ask me for their locations.

# Steps

## 1. Checkout the version of the opentelemetry-go repo that matches the version that we are upgrading to.
- Navigate to the local opentelemetry-go repo and checkout the version that matches the version that we are upgrading to.
- Navigate to the local opentelemetry-go-contrib repo and checkout the version that matches the version that we are upgrading to.
- If you cannot figure out the path to the local repos, ask me for their locations.
- Use these local repos as a reference in the rest of the steps.

## 2. Update versions in go.mod
- Update the `go.mod` versions with the appropriate otel go and otel go contrib dependency versions. Dependencies prefixed with `go.opentelemetry.io/otel` correspond to the repo `github.com/open-telemetry/opentelemetry-go` and dependencies prefixed with `go.opentelemetry.io/contrib` correspond to the repo `github.com/open-telemetry/opentelemetry-go-contrib`.
- Run `go mod tidy` from the root directory to update the `go.sum` file.

## 3. Update `go.mod` versions across all modules
Run from repo root:

```sh
make tidy
```
This will update all `go.mod` files in the repo to the latest compatible versions of OTEL deps.

Notes:
- If `go mod tidy` complains about Go version constraints, adjust the `-go=` value.

## 4. Update Traceable custom `batchspanprocessor`
This repo vendors/adapts OTEL SDK batch span processor internals here:

- `hypertrace/goagent/instrumentation/opentelemetry/batchspanprocessor/`

**NOTE**: When executing this step, make sure to use the local repo of `github.com/open-telemetry/opentelemetry-go` as the opentelemetry-go repo.

Compare the original opentelemetry-go file to the corresponding file suffixed with `.go_original` and apply the changes to the file with our customizations to the file suffixed with `.modified.go`. For example for `batch_span_processor.go` the original would be `batch_span_processor.go_original` and `batch_span_processor.go_modified.go` would contain our customizations. Once the changes are done copy the opentelemetry-go file to the one ending with `.go_original` to preserve the original file and help with future updates. Do this for these files in the `hypertrace/goagent/instrumentation/opentelemetry/batchspanprocessor/` directory:

- `batch_span_processor.go` whose opentelemetry-go original is `sdk/trace/batch_span_processor.go`
- `env.go` whose opentelemetry-go original is `sdk/trace/internal/env/env.go`
- `observ/batch_span_processor.go` whose opentelemetry-go original is `sdk/trace/internal/observ/batch_span_processor.go`

If the opentelemetry-go files have been moved, try to look for where they moved to and update this workflow to reflect the new paths.

For these 2 files, copy them as they are from the opentelemetry-go repo but maintain our modifications which for now is the first line that starts with `// Adapted from ...` followed by the path to upstream file.
- `internal/x/x.go` whose opentelemetry-go original is `sdk/internal/x/x.go`
- `internal/x/features.go` whose opentelemetry-go original is `sdk/internal/x/features.go`

## 5. Fix compile breaks due to OTEL API changes
Common areas:
- Metric/log SDK package versioning (0.xx vs 1.xx)
- Contrib instrumentation API changes

## 6. Verify
From repo root:

```sh
go test -count=1 ./...
make check-examples
```

If repo tests are known-broken (per doc), at minimum run:
- `go test -count=1 ./...` in root
- `go test -count=1 ./...` in key nested modules that import OTEL
