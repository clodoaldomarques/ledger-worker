# Tarefas — Spec 001

**Legenda:** `[x]` verificado no código; `[ ]` lacuna ou trabalho pendente.

## Validação da spec

- [x] Confirmar que `message.NewConsumer` passa `config.New()` e `Handler` ao core-sdk.
- [x] Confirmar o mapeamento de `correlation_id`, `org_id`, `program.id` e `account_id`.
- [x] Confirmar `Producer = "regular"` e o uso de `decimal.Decimal` em `Amounts`/`Fees`.
- [x] Confirmar os headers `x-cid`, `x-tenant` e `Content-Type`.
- [x] Confirmar que somente `201` é tratado como sucesso.
- [x] Confirmar `ReadyToTrip: ConsecutiveFailures > 5` e `Timeout: 15s`.
- [x] Confirmar que falha do handler não deleta a mensagem e que a DLQ usa `ApproximateReceiveCount >= 10`.
- [x] Confirmar os três spans OTel e o log estruturado em caso de erro.
- [x] Executar `go test ./...` sem serviços externos.

## Lacunas conhecidas

- [ ] Validar campos obrigatórios antes de desreferenciar `ProcessingCode` e `amount[].description`, substituindo pânico por erro de mensagem inválida.
- [ ] Evitar perda de precisão: ler valores decimais como string ou `json.Number` em vez de `float64`.
- [ ] Chamar `LedgerEventsApi.Close()` no graceful shutdown.
- [ ] Definir política de idempotência para redelivery (não existe hoje).
- [ ] Definir retry/backoff para a API downstream, hoje limitado a visibility timeout + DLQ.
- [ ] Adicionar teste de `buildLedgerEvent`, que atualmente não tem cobertura direta.
- [ ] Adicionar expectation de `POST /v1/ledger/events` no mockserver para permitir teste ponta a ponta (ver spec 003).
