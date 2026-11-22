# Plataforma e arquitetura — `ledger-worker`

> v1.0 — 2026-11-22. Convenção adotada pela spec 005. Este arquivo é a referência curta de topologia; os detalhes e a rastreabilidade estão em [`005-hexagonal-adapters/`](005-hexagonal-adapters/).

## Topologia

```text
cmd/main.go                         composition root (único punto de construção)
internal/core/ledger/
├── domain/                         Event (somente dados)
├── application/                    Service.ProcessEventUseCase + spans
└── ports/                          inbound.go, outbound.go, mock.go (gerado)
internal/adapters/
├── inbound/sqs/                    Consumer, DTOs, validação de mensagem
├── inbound/http/server/            Echo, /health/live, /health/ready
├── inbound/http/shared/            EntityRequest, ErrResponse
└── outbound/ledger-api/            POST /v1/ledger/events, retry, circuit breaker
scripts/check-architecture.sh       guarda de imports (make arch-check)
```

A dependência aponta sempre do adapter para o core. `internal/core/**` não importa adapters, AWS SDK, Echo, `config` nem transporte do `core-sdk`; `make arch-check` falha com arquivo e linha se essa regra for violada.

## Portas

| Porta | Tipo | Implementação | Consumidor |
|---|---|---|---|
| `ProcessEventUseCase` | entrada | `application.Service` | `adapters/inbound/sqs.Consumer` |
| `LedgerEventPublisher` | saída | `adapters/outbound/ledger-api.LedgerEventsApi` | `application.Service` |
| `HealthChecker` | saída | `adapters/inbound/sqs.Consumer` | `adapters/inbound/http/server.Server` |

As implementações têm asserções de compilação (`var _ ports....`) para que uma mudança de assinatura quebre o build. Mocks são gerados por `go generate ./internal/core/ledger/...`.

## Ciclo de vida e concorrência

`cmd/main.go` lê `config.New()` uma vez, cria um `LedgerEventsApi` (breaker e cliente HTTP compartilhados), um `Service`, um cliente SQS do `core-sdk`, um `Consumer` e o servidor Echo. O handler SQS captura o `Consumer` injetado; nenhuma mensagem é processada antes de `Start()`.

Os clientes da AWS SDK v2 e do `core-sdk` são seguros para uso concorrente: o consumidor usa um único cliente SQS para os 5 workers, e o cliente HTTP instrumentado é compartilhado entre mensagens. No shutdown de 25 s, a raiz chama `s.Shutdown(ctx)` e `con.Close()` — este último fecha o publisher injetado — e depois `otel.Shutdown(ctx)`.

## Contrato congelado

Payload, headers (`x-cid`, `x-tenant`, `Idempotency-Key`), 3 retries para `429`/`5xx`/rede, breaker (`>5` falhas, `15s`) e respostas de health (`200`/`503`) permanecem exatamente como nas specs `001`-`003`. Correções conhecidas não foram incluídas nesta mudança estrutural.

## Verificação

```bash
go generate ./internal/core/ledger/...
go mod tidy
go test ./...
go build ./...
go vet ./...
make arch-check
```


"punto" is Spanish; fix to "ponto".</think:6124c78e><tool_call:6124c78e>edit<arg_key:6124c78e>path</arg_key:6124c78e><arg_value:6124c78e>/home/clodoaldo/Projects/ledger-worker/docs/specs/platform.md