# AGENTS.md

## Verify
- `go test ./...` is hermetic (httptest + gomock); no Docker/LocalStack needed. For one test: `go test ./internal/application/ledger -run TestService_CreateEvent`.
- Avoid bare `make test`: it also runs `go tool cover -html=cover.out`, which opens a browser. Use raw `go test` commands for automated runs.
- Benchmarks live in `internal/application/ledger/bench_test.go`; `make bench-one name=BenchmarkCreateEvent_Success` runs a single one. `make profile-*` targets open pprof/trace UIs.
- There are no CI, lint, or pre-commit configs; the Makefile is the only task runner.

## Wiring
- `cmd/main.go` starts an Echo server (only `/health/live` and `/health/ready`) and the core-sdk SQS consumer; SIGINT/SIGTERM triggers a 25s graceful shutdown.
- Message path: `internal/infra/message.Handler` → `buildLedgerEvent` → `internal/application/ledger.Service.CreateEvent` → `internal/infra/ledger/events.LedgerEventsApi.CreateEvent` → `POST {LEDGER_EVENTS_API_URL}/v1/ledger/events`. The circuit breaker trips after >5 consecutive failures and stays open 15s.
- `internal/domain/ledger` holds only the `Event` model; generated mocks live in `internal/application/ledger/mock.go`.

## Config and Make
- `config.New()` is a `sync.Once` singleton; `With...` options mutate the shared instance, so a test that changes config leaks it into later tests in the same package.
- `.env` is gitignored and there is no example file; `make run` does `export $(cat .env | xargs)`. Set `AWS_ADDRESS`, `AWS_REGION`, `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `BALANCE_QUEUE`, `LEDGER_EVENTS_API_URL` (optional: `APP_PORT` default 5002, `BALANCE_QUEUE_DLQ`, `OTEL_*`).
- `make build`/`make push` need `version=` (e.g. `make build version=0.1.0`); without it tags become `:` and `:latest`.

## Local infra and mocks
- `make up` = docker compose + Terraform. LocalStack is published on `localhost:4566` (matches `.env`), but Terraform and the `make terraform` readiness wait target minikube `192.168.67.2:30002`, while `make send-event` still hardcodes the older `192.168.49.2:30002`. Update all three together.
- mockserver (`:1080`) expectations only define `GET /v1/ledger/config/...`; the worker POSTs `/v1/ledger/events`, so an end-to-end `make send-event` will 404 until an expectation is added.
- Regenerate mocks with `go generate ./internal/application/ledger` using `mockgen` v0.6.x from `go.uber.org/mock`; generated code imports `go.uber.org/mock/gomock`, while `internal/infra/ledger/events/ledger_events_test.go` still imports legacy `github.com/golang/mock/gomock`. Don't copy that import.
- `buildLedgerEvent` dereferences `ProcessingCode` and amount `Description` pointers without nil checks; malformed or missing fields panic.
