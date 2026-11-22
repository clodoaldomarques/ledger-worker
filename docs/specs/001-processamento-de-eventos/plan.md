# Plano técnico — Spec 001

> Este plano documenta a solução **já implementada**. Não é um plano de trabalho futuro.

## Contexto técnico

- Go `1.26.1` em `go.mod`; a imagem Docker compila com Go `1.27.1`.
- `core-sdk v0.8.8` fornece consumidor SQS, cliente HTTP instrumentado, OTel e logger.
- `gobreaker v1.0.0` protege a integração HTTP; `shopspring/decimal v1.5.0` representa valores.
- Entrada: `cmd/main.go` inicia OTel, servidor HTTP e consumidor; o shutdown usa contexto de 25 s.

## Decisões técnicas atuais

| Área | Decisão | Local |
|---|---|---|
| Consumo | 5 workers, long polling 20 s, visibility 30 s, lote de 10 (padrões do core-sdk) | `core-sdk/pkg/aws/sqs/consumer.go` |
| Transformação | `json.Unmarshal` direto em `TransactionCreated`, depois mapeamento manual | `internal/infra/message/consumer.go` |
| Domínio | `Event` sem methods, apenas dados | `internal/domain/ledger/event.go` |
| Orquestração | `Service.CreateEvent` só adiciona tracing e delega | `internal/application/ledger/service.go` |
| Abstração | `EventsProvider` com mock gerado por `go:generate` | `internal/application/ledger/interfaces.go` |
| Integração | POST com `gobreaker`, headers de tenant/cid | `internal/infra/ledger/events/ledger_events.go` |
| DLQ | Movimento manual pelo core-sdk quando `ApproximateReceiveCount >= 10` | `core-sdk/pkg/aws/sqs/consumer.go` |

## Fluxo de execução

```text
SQS balance-sqs-queue
  → sqs.Consumer (core-sdk, 5 workers)
  → message.Handler
      → tracer.NewSpanFromContext("Message::Handler")
      → buildLedgerEvent (TransactionCreated → ledger.Event)
      → service.CreateEvent
          → tracer.NewSpanFromContext("Service::Handler")
          → LedgerEventsApi.CreateEvent
              → tracer.NewSpanFromContext("LedgerEventsApi::CreateEvent")
              → circuitBreaker.Execute
                  → POST {LEDGER_EVENTS_API_URL}/v1/ledger/events
  → sucesso: DeleteMessage | falha: manter na fila | 10 recebimentos: DLQ
```

## Transformação de dados

```text
correlation_id ──────────────→ Cid ──────────────────→ header x-cid
org_id ─────────────────────→ OrgID ─────────────────→ header x-tenant
program.id ─────────────────→ ProgramID ─────────────→ corpo program_id
account_id ─────────────────→ AccountID ─────────────→ corpo account_id
processing_code ────────────→ ProcessingCode ────────→ corpo processing_code
"regular" (constante) ──────→ Producer ──────────────→ corpo producer
amount[].description/value → Amounts map ───────────→ corpo amounts
tax[].type/value ───────────→ Fees map ──────────────→ corpo fees
```

## Configuração relevante

| Variável | Uso na spec 001 |
|---|---|
| `BALANCE_QUEUE` | URL da fila de entrada. |
| `BALANCE_QUEUE_DLQ` | URL da DLQ usada pelo core-sdk; vazia desativa o movimento. |
| `AWS_*` | Endpoint, região e credenciais do cliente SQS. |
| `LEDGER_EVENTS_API_URL` | Base URL da API downstream. |

`config.MaxReceiveCount()` retorna `10` fixo. Não existe variável para ajustar workers, visibility timeout ou limites do circuit breaker.

## Verificação

- Testes de unidade: `go test ./internal/application/ledger/...` e `go test ./internal/infra/ledger/events/...`.
- Teste focado: `go test ./internal/application/ledger -run TestService_CreateEvent`.
- Benchmark: `make bench-one name=BenchmarkCreateEvent_Success`.
- Integração local depende de LocalStack, mockserver e alinhamento dos endpoints (spec 003).

## Riscos e dívida técnica

1. Pânico em mensagens sem `processing_code` ou `amount[].description`.
2. Perda de precisão monetária pela conversão via `float64`.
3. Ausência de idempotência: redelivery gera novos POSTs.
4. `Close()` do cliente HTTP não é chamado no shutdown.
5. Dependência de um singleton de config mutável dificulta testes isolados.

## Cobertura dos requisitos

| Requisito | Evidência |
|---|---|
| RF-01, RF-07 | `internal/infra/message/consumer.go` + core-sdk |
| RF-02 | `buildLedgerEvent`, `buildAmounts`, `buildFees` |
| RF-03 a RF-06 | `internal/infra/ledger/events/ledger_events.go`, `types.go` |
| RF-08 | `message.Handler`, `application.Service`, `LedgerEventsApi` |
