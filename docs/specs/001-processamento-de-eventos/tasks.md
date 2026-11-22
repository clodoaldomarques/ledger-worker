# Tarefas — Spec 001

**Legenda:** `[x]` verificado no código; `[ ]` lacuna ou trabalho pendente.

## Lacunas originais — corrigidas

- [x] Validar `processing_code` e `amount[].description` antes de desreferenciá-los, retornando `ErrInvalidMessage` em vez de pânico.
- [x] Evitar perda de precisão: `json.Number` + `decimal.NewFromString` substituíram `float64`.
- [x] Chamar `LedgerEventsApi.Close()` no graceful shutdown.
- [x] Definir política de idempotência: header `Idempotency-Key` com o `correlation_id`.
- [x] Definir retry/backoff: 3 tentativas para `429`/`5xx`/erro de rede, com backoff exponencial e jitter.
- [x] Adicionar teste de `buildLedgerEvent`, cobrindo conversão, precisão e entradas inválidas.
- [x] Adicionar expectation de `POST /v1/ledger/events` no mockserver.

## Correção adicional descoberta durante a implementação

- [x] Compartilhar o circuit breaker entre mensagens: `events.New(ctx)` saiu do handler e passou a ser chamado uma vez em `message.NewConsumer`.
- [x] Evitar criar um consumidor SQS por requisição de readiness injetando o consumidor no servidor (spec 002).

## Validação da spec

- [x] `go test ./...` passa sem Docker, LocalStack ou mockserver.
- [x] `go vet ./...` passa.
- [x] Teste confirma que `400` não é retentado.
- [x] Teste confirma que `5xx` é retentado até `201`.
- [x] Teste confirma que o breaker abre após mais de cinco execuções falhas e bloqueia a próxima.
- [x] Teste confirma `x-cid`, `x-tenant` e `Idempotency-Key`.
- [x] `expectations.json` é um JSON válido.

## Lacunas restantes

- [ ] Deduplicação depende da API downstream; considerar dedupe local com Redis se a API não garantir idempotência.
- [ ] Tornar tentativas/delays do retry configuráveis por ambiente.
- [ ] Adicionar métricas para retry, circuit breaker e mensagens inválidas.
- [ ] Reduzir o atributo de span `event_type` para campos selecionados em vez da struct inteira.
