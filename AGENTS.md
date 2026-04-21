# AGENTS.md - Go Agent

## Project Overview

Go Agent provides instrumentation features for collecting tracing data and securing applications by selectively blocking requests using Traceable features. It includes OpenTelemetry-based tracing instrumentation for HTTP servers, gRPC, and other Go frameworks, along with a Traceable filter for request security.

This is a Go modules-based project with support for multiple Linux distributions and Alpine/Debian/Ubuntu/CentOS environments.

## Build System

- **Build tool**: Go modules
- **Build**: `go build ./...`
- **Build with Traceable filter**: `go build -tags 'traceable_filter' -o /path-to-app/myapp`
- **Install dependencies**: `go get -v -t -d ./...` or `make deps`
- **Tidy dependencies**: `make tidy`

## Testing

- **Run all tests**: `go test ./...`
- **Run tests with coverage**: `go test -count=1 -v -race -cover ./...` or `make test`
- **Run benchmarks**: `make bench`
- **Run specific test**: `go test -run TestName ./path/to/package`

## Linting & Formatting

- **Format code**: `gofmt -w -s ./` or `make fmt`
- **Run linters**: `make lint` (requires golangci-lint)
- **Check vanity imports**: `make check-vanity-import`
- **Run before committing**: Always run `make fmt && make test` before creating commits

## Git Workflow

- **Branch naming**: `JIRA-TICKET-short-description` or `NO-TICKET-short-description`
- **Commit format**: `JIRA-TICKET: Description` or `NO-TICKET: Description`
- **PR title format**: Same as commit format
- **Default branch**: `main`

## Technology Stack

- **Language**: Go (1.21+)
- **Build Tool**: Go modules
- **Tracing**: OpenTelemetry
- **Instrumentation**: HTTP, gRPC, database clients
- **Security**: Traceable filter for request blocking
- **Supported Platforms**: Linux (Debian, Ubuntu, CentOS, Alpine, Amazon Linux)

## Key Features

- **Automatic Instrumentation**: HTTP server/client, gRPC, database clients
- **OpenTelemetry Integration**: Standard tracing and metrics collection
- **Traceable Filter**: Security filter for blocking malicious requests
- **Configuration**: File-based, environment variables, or code-based config
- **Multi-Platform**: Support for various Linux distributions
- **Propagation**: Standard context propagation for distributed tracing

## DOs

- Follow existing code patterns in the codebase
- Run `make fmt` before committing to ensure code style compliance
- Run `make test` before committing to ensure tests pass
- Use descriptive commit messages with ticket references (or NO-TICKET prefix)
- Add unit tests for all new business logic
- Run `make lint` to check for common issues
- Update examples when adding new instrumentation features

## DON'Ts

- Never force push to `main`
- Never commit secrets, `.env` files, or credentials
- Never skip git hooks (`--no-verify`)
- Never run destructive commands without confirmation
- Don't commit without running `make fmt` (code formatting)
- Don't commit without running `make test` (tests must pass)
- Don't add new dependencies without checking compatibility across supported Go versions

## Commands to Never Run

- `git push --force origin main`
- `git commit --no-verify` or `git push --no-verify`
- `rm -rf /` (or any destructive recursive delete)

## Important Patterns

### Build Tags

The Traceable filter requires the `traceable_filter` build tag:
```bash
go build -tags 'traceable_filter' -o myapp
```

### Configuration

Config values can be declared in:
- Config files
- Environment variables
- Code (using `config.Load()`)

See `config/README.md` for detailed configuration options.

### Instrumentation

The agent provides automatic instrumentation for:
- HTTP servers and clients (net/http)
- gRPC servers and clients
- Database clients (SQL, MongoDB, etc.)

Import the appropriate instrumentation package and follow existing patterns.

## Development Workflow

1. Create a branch following the naming convention
2. Make your changes
3. Run `make fmt` to format code
4. Run `make test` to verify tests pass
5. Run `make lint` to check for issues
6. Commit with proper message format
7. Create a PR with the same title format

## Additional Resources

- See parent repository (activity-event-service) for detailed Go coding standards
- Traceable agent config proto: https://github.com/Traceableai/agent-config
- OpenTelemetry Go: https://opentelemetry.io/docs/languages/go/
