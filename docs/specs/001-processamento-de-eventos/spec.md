# Spec 001 — Processamento de eventos SQS → Ledger Events API

**Status:** Implementado (reverse-engineered)  
**Fonte:** `cmd/main.go`, `internal/infra/message/*`, `internal/application/ledger/*`, `internal/domain/ledger/*`, `internal/infra/ledger/events/*`, `core-sdk v0.8.8`  
**Commit de referência:** `5794727`

## Visão geral

O worker consome mensagens da fila `BALANCE_QUEUE`, converte o JSON de transação em um `ledger.Event` e envia um `EventRequest` para `POST {LEDGER_EVENTS_API_URL}/v1/ledger/events`. O consumidor do `core-sdk` usa 5 workers concorrentes, long polling de 20 s, visibility timeout de 30 s e até 10 mensagens por recebimento — valores padrão da biblioteca, não sobrescritos por este projeto.

## Cenários de uso

### US1 — Processar uma transação válida

**Given** uma mensagem SQS com `processing_code`, `amount[].description` e os demais campos usados pela conversão presentes  
**When** `message.Handler` recebe a mensagem  
**Then** um `ledger.Event` é construído e `Service.CreateEvent` delega ao `EventsProvider`  
**And** se a API responder `201`, o core-sdk deleta a mensagem da fila e o handler retorna `nil`.

### US2 — Falhar antes de chamar a API quando o JSON é inválido

**Given** o corpo da mensagem não desserializa em `TransactionCreated`  
**When** `buildLedgerEvent` retorna erro  
**Then** o erro é registrado no span `Message::Handler` e o handler retorna erro  
**And** nenhuma requisição HTTP é feita; a mensagem permanece na fila para nova tentativa.

### US3 — Propagar falha da API downstream

**Given** a API responde status diferente de `201` (por exemplo, `400`)  
**When** `LedgerEventsApi.CreateEvent` executa  
**Then** retorna erro no formato `api error: status <n>, body: <corpo>`  
**And** o span é marcado com erro, log estruturado é emitido e a mensagem não é deletada.

### US4 — Falha rápida com circuit breaker

**Given** mais de cinco falhas consecutivas na chamada à API  
**When** outra mensagem chega dentro dos 15 segundos seguintes  
**Then** o circuito está `OPEN` e `Execute` falha sem abrir nova requisição HTTP.

### US5 — Encaminhar para DLQ

**Given** `BALANCE_QUEUE_DLQ` configurada e `ApproximateReceiveCount >= 10`  
**When** a mensagem é recebida  
**Then** o core-sdk envia o corpo para a DLQ e deleta a mensagem da fila original, sem executar o handler.

## Requisitos funcionais

- **RF-01** Criar o consumidor com `sqs.NewConsumer(ctx, config.New(), Handler)` a partir de `message.NewConsumer`.
- **RF-02** Converter `TransactionCreated` em `ledger.Event` preenchendo `Cid`, `OrgID`, `ProgramID`, `AccountID`, `ProcessingCode`, `Producer`, `Amounts` e `Fees`.
- **RF-03** Enviar `EventRequest` com `processing_code`, `program_id`, `account_id`, `producer`, `amounts` e `fees`.
- **RF-04** Definir os headers `Content-Type: application/json`, `x-cid` e `x-tenant` na requisição.
- **RF-05** Aceitar somente `201 Created` como sucesso; qualquer outro status é erro.
- **RF-06** Proteger a chamada HTTP com circuit breaker nomeado `LedgerEventsAPI`.
- **RF-07** Não deletar mensagens cujo handler retornou erro; mover para DLQ após 10 recebimentos quando a DLQ estiver configurada.
- **RF-08** Emitir spans e logs estruturados com identificadores da mensagem e do evento.

## Entidades e contratos

### Entrada — `TransactionCreated` (`internal/infra/message/types.go`)

| Campo JSON | Tipo | Destino no `ledger.Event` |
|---|---|---|
| `correlation_id` | `string` | `Cid` |
| `org_id` | `string` | `OrgID` |
| `program.id` | `int64` | `ProgramID` |
| `account_id` | `int64` | `AccountID` |
| `processing_code` | `*string` | `ProcessingCode` (desreferenciado) |
| `amount[].description` | `*string` | chave de `Amounts` (desreferenciada) |
| `amount[].value` | `float64` | valor de `Amounts` via `decimal.NewFromFloat` |
| `tax[].type` | `string` | chave de `Fees` |
| `tax[].value` | `float64` | valor de `Fees` via `decimal.NewFromFloat` |

`Producer` é fixo: `"regular"`. Outros campos do payload são decodificados, mas não usados na conversão.

### Domínio — `ledger.Event`

`Cid`, `OrgID`, `ProgramID`, `AccountID`, `ProcessingCode`, `Producer`, `Amounts map[string]decimal.Decimal`, `Fees map[string]decimal.Decimal`.

### Saída — `EventRequest` (`POST /v1/ledger/events`)

`processing_code`, `program_id`, `account_id`, `producer`, `amounts`, `fees`. `Cid` e `OrgID` **não** vão no corpo; seguem nos headers `x-cid` e `x-tenant`. O corpo é serializado com `encoding/json`.

## Requisitos não funcionais

- **RNF-01** Valores monetários são `decimal.Decimal` no domínio e na requisição.
- **RNF-02** Três spans OTel encadeados: `Message::Handler` → `Service::Handler` → `LedgerEventsApi::CreateEvent`.
- **RNF-03** Cliente HTTP instrumentado pelo core-sdk (`httpclient.Client()`).
- **RNF-04** Circuit breaker: `ConsecutiveFailures > 5`, `Timeout = 15s`, transições logadas.
- **RNF-05** Processamento concorrente e DLQ são fornecidos pelo core-sdk; este repositório apenas fornece a configuração.

## Critérios de aceite

- `go test ./internal/application/ledger -run TestService_CreateEvent` passa.
- `go test ./internal/infra/ledger/events` passa, cobrindo respostas `201` e `400`.
- O teste com `httptest` valida os headers `Content-Type`, `x-cid` e `x-tenant`.
- O benchmark `BenchmarkCreateEvent_Success` continua disponível em `bench_test.go`.

## Limitações conhecidas e fora de escopo

- `ProcessingCode` e `amount[].description` são ponteiros desreferenciados sem verificação de `nil`: mensagens sem esses campos causam pânico.
- `decimal.NewFromFloat` converte um `float64` já impreciso; não há leitura via `json.Number` ou string decimal.
- O atributo de span `event_type` recebe a struct `ledger.Event` inteira.
- `LedgerEventsApi.Close()` existe, mas não é chamado no graceful shutdown.
- Não há idempotência nem retry/backoff no worker; retries dependem da visibilidade da SQS e da DLQ. O README lista idempotência e retry como melhorias futuras.
- O mockserver local não define `POST /v1/ledger/events` (ver spec 003), então um teste ponta a ponta local retorna `404`.
