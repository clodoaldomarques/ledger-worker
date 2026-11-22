# Plano técnico — Spec 001

> Este plano documenta a solução **implementada**, incluindo as correções das lacunas originais.

## Contexto técnico

- Go `1.26.1` em `go.mod`; a imagem Docker compila com Go `1.27.1`.
- `core-sdk v0.8.8` fornece consumidor SQS, cliente HTTP instrumentado, OTel e logger.
- `gobreaker v1.0.0` protege a integração HTTP; `shopspring/decimal v1.5.0` representa valores.
- `math/rand/v2` gera jitter do backoff; não há dependência nova de retry.

## Decisões técnicas atuais

| Área | Decisão | Local |
|---|---|---|
| Consumo | 5 workers, long polling 20 s, visibility 30 s, lote de 10 (padrões do core-sdk) | `core-sdk/pkg/aws/sqs/consumer.go` |
| Ciclo de vida | `LedgerEventsApi` e circuit breaker criados na raiz e injetados; fechados por `Consumer.Close()` | `cmd/main.go`, `internal/adapters/inbound/sqs/consumer.go` |
| Validação | `ErrInvalidMessage` para JSON inválido e campos obrigatórios ausentes | `internal/adapters/inbound/sqs/errors.go` |
| Precisão | `json.Number` + `decimal.NewFromString`, sem conversão por `float64` | `types.go`, `consumer.go` |
| Domínio | `Event` sem methods, apenas dados | `internal/core/ledger/domain/event.go` |
| Orquestração | `Service.CreateEvent` implementa `ProcessEventUseCase`, adiciona tracing e delega ao publisher | `internal/core/ledger/application/service.go` |
| Abstração | `LedgerEventPublisher` com mock gerado por `go:generate` | `internal/core/ledger/ports/outbound.go`, `mock.go` |
| Integração | POST com breaker compartilhado, headers de tenant/cid/idempotência | `internal/adapters/outbound/ledger-api/ledger_events.go` |
| Retry | 3 tentativas, 100 ms base, teto 1 s, jitter 0,5–1,5, somente 429/5xx/rede | `ledger_events.go` |
| DLQ | Movimento manual pelo core-sdk quando `ApproximateReceiveCount >= 10` | `core-sdk/pkg/aws/sqs/consumer.go` |

## Fluxo de execução

```text
SQS balance-sqs-queue
  → sqs.Consumer (core-sdk, 5 workers)
  → Consumer.Handle
      → tracer.NewSpanFromContext("Message::Handler")
      → buildLedgerEvent (valida → TransactionCreated → ledger.Event)
      → Service.CreateEvent
          → tracer.NewSpanFromContext("Service::Handler")
          → LedgerEventsApi.CreateEvent
              → tracer.NewSpanFromContext("LedgerEventsApi::CreateEvent")
              → circuitBreaker.Execute (instância única)
                  → até 3× POST {LEDGER_EVENTS_API_URL}/v1/ledger/events
  → sucesso: DeleteMessage | falha: manter na fila | 10 recebimentos: DLQ
```

## Transformação de dados

```text
correlation_id ──────────────→ Cid ──────────────────→ x-cid + Idempotency-Key
org_id ─────────────────────→ OrgID ─────────────────→ x-tenant
program.id ─────────────────→ ProgramID ─────────────→ corpo program_id
account_id ─────────────────→ AccountID ─────────────→ corpo account_id
processing_code ────────────→ ProcessingCode ────────→ corpo processing_code
"regular" (constante) ──────→ Producer ──────────────→ corpo producer
amount[].description/value → Amounts map ───────────→ corpo amounts
tax[].type/value ───────────→ Fees map ──────────────→ corpo fees
```

## Retry e idempotência

- `isRetryable`: `429` ou `>= 500`; erros de `httpClient.Do` também são retentáveis.
- `4xx` diferentes de `429` falham imediatamente para não reproduzir erros de contrato.
- O backoff usa `time.NewTimer` e `select` com `ctx.Done()`, respeitando shutdown.
- `Idempotency-Key` usa o mesmo `correlation_id` de `x-cid`; a deduplicação é responsabilidade da API downstream.

## Configuração relevante

| Variável | Uso na spec 001 |
|---|---|
| `BALANCE_QUEUE` | URL da fila de entrada. |
| `BALANCE_QUEUE_DLQ` | URL da DLQ usada pelo core-sdk; vazia desativa o movimento. |
| `AWS_*` | Endpoint, região e credenciais do cliente SQS. |
| `LEDGER_EVENTS_API_URL` | Base URL da API downstream. |

Número de tentativas, delays e limites do circuit breaker são constantes no código, não variáveis de ambiente.

## Verificação

- `go test ./internal/adapters/inbound/sqs` — conversão, precisão, validação e uso da porta de entrada.
- `go test ./internal/adapters/outbound/ledger-api` — headers, retry, breaker e idempotência.
- `go test ./internal/core/ledger/application` — serviço, mocks e fake da porta de saída.
- `make bench-one name=BenchmarkCreateEvent_Success` — benchmark do serviço.

## Riscos e dívida técnica

1. Idempotência depende da API downstream; o worker não detecta duplicatas localmente.
2. Retry em processo ocupa brevemente o worker; delays são fixos.
3. Atributo de span `event_type` serializa a struct inteira.
4. Sem métricas específicas para retry/breaker/mensagens inválidas.
5. Config continua sendo um singleton mutável.

## Cobertura dos requisitos

| Requisito | Evidência |
|---|---|
| RF-01, RF-09 | `internal/adapters/inbound/sqs/consumer.go` + core-sdk (cliente criado em `cmd/main.go`) |
| RF-02, RF-03 | `buildLedgerEvent`, `parseDecimal`, `ErrInvalidMessage` |
| RF-04 a RF-08, RF-11 | `internal/adapters/outbound/ledger-api/ledger_events.go` |
| RF-10 | `sqs.Consumer.Handle`, `application.Service`, `LedgerEventsApi` |
