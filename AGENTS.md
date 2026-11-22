# AGENTS.md

## Verify
- `go test ./...` is hermetic (httptest + gomock); no Docker/LocalStack needed. For one test: `go test ./internal/application/ledger -run TestService_CreateEvent`.
- Avoid bare `make test`: it also runs `go tool cover -html=cover.out`, which opens a browser. Use raw `go test` commands for automated runs.
- Benchmarks live in `internal/application/ledger/bench_test.go`; `make bench-one name=BenchmarkCreateEvent_Success` runs a single one. `make profile-*` targets open pprof/trace UIs.
- There are no CI, lint, or pre-commit configs; the Makefile is the only task runner.

## Wiring
- `cmd/main.go` starts an Echo server (only `/health/live` and `/health/ready`) and the core-sdk SQS consumer; SIGINT/SIGTERM triggers a 25s graceful shutdown.
- Message path: `message.Consumer.handle` → `buildLedgerEvent` → `application/ledger.Service.CreateEvent` → `infra/ledger/events.LedgerEventsApi.CreateEvent` → `POST {LEDGER_EVENTS_API_URL}/v1/ledger/events`.
- The API client and circuit breaker are created once in `message.NewConsumer`, shared across messages and closed on shutdown; breaker trips after >5 failures (15s), with 3 retries on `429`/`5xx`/network errors and `Idempotency-Key` sent as the correlation id.
- `internal/domain/ledger` holds only the `Event` model; generated mocks live in `internal/application/ledger/mock.go`.

## Config and Make
- `config.New()` is a `sync.Once` singleton; `With...` options mutate the shared instance, so a test that changes config leaks it into later tests in the same package.
- `.env` is gitignored and there is no example file; `make run` does `export $(cat .env | xargs)`. Set `AWS_ADDRESS`, `AWS_REGION`, `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `BALANCE_QUEUE`, `LEDGER_EVENTS_API_URL` (optional: `APP_PORT` default 5002, `BALANCE_QUEUE_DLQ`, `OTEL_*`).
- `make build`/`make push` need `version=` (e.g. `make build version=0.1.0`); without it tags become `:` and `:latest`.

## Local infra and mocks
- `make up` = docker compose + Terraform. LocalStack is published on `localhost:4566` (matches `.env`), but Terraform and the `make terraform` readiness wait target minikube `192.168.67.2:30002`, while `make send-event` still hardcodes the older `192.168.49.2:30002`. Update all three together.
- mockserver (`:1080`) now defines `POST /v1/ledger/events` → 201 besides `GET /v1/ledger/config/...`; the remaining local blocker is the stale endpoint hardcoded in `make send-event`.
- Regenerate mocks with `go generate ./internal/application/ledger` using `mockgen` v0.6.x from `go.uber.org/mock`; generated code imports `go.uber.org/mock/gomock`, while `internal/infra/ledger/events/ledger_events_test.go` still imports legacy `github.com/golang/mock/gomock`. Don't copy that import.
- `buildLedgerEvent` validates `ProcessingCode` and amount `Description` and parses values via `json.Number` + `decimal.NewFromString`; malformed messages return `ErrInvalidMessage` instead of panicking.
