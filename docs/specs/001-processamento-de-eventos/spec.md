# Spec 001 — Processamento de eventos SQS → Ledger Events API

**Status:** Implementado (reverse-engineered)  
**Fonte:** `cmd/main.go`, `internal/infra/message/*`, `internal/application/ledger/*`, `internal/domain/ledger/*`, `internal/infra/ledger/events/*`, `core-sdk v0.8.8`  
**Commit de referência:** `5794727` (spec original); comportamento atual atualizado após correções

## Visão geral

O worker consome mensagens da fila `BALANCE_QUEUE`, valida e converte o JSON de transação em um `ledger.Event` e envia um `EventRequest` para `POST {LEDGER_EVENTS_API_URL}/v1/ledger/events`. O cliente da API e o circuit breaker são criados **uma única vez** em `message.NewConsumer`, compartilhados entre mensagens e fechados no graceful shutdown.

O consumidor do `core-sdk` usa 5 workers concorrentes, long polling de 20 s, visibility timeout de 30 s e até 10 mensagens por recebimento — valores padrão da biblioteca, não sobrescritos por este projeto.

## Cenários de uso

### US1 — Processar uma transação válida

**Given** uma mensagem SQS com `processing_code`, `amount[].description` e valores decimais válidos  
**When** `message.Consumer.handle` recebe a mensagem  
**Then** um `ledger.Event` é construído e `Service.CreateEvent` delega ao `EventsProvider`  
**And** se a API responder `201`, o core-sdk deleta a mensagem da fila e o handler retorna `nil`.

### US2 — Rejeitar mensagem inválida antes de chamar a API

**Given** o corpo não é JSON, falta `processing_code` ou falta `amount[].description`  
**When** `buildLedgerEvent` falha  
**Then** o erro é `ErrInvalidMessage`, registrado no span `Message::Handler`, e o handler retorna erro  
**And** nenhuma requisição HTTP é feita; a mensagem permanece na fila para nova tentativa.

### US3 — Propagar falha da API downstream

**Given** a API responde status `4xx` diferente de `429`  
**When** `LedgerEventsApi.CreateEvent` executa  
**Then** retorna erro no formato `api error: status <n>, body: <corpo>` sem retry  
**And** o span é marcado com erro, log estruturado é emitido e a mensagem não é deletada.

### US4 — Retentar falhas transitórias

**Given** a API responde `429`/`5xx` ou ocorre erro de rede  
**When** a chamada falha nas duas primeiras tentativas  
**Then** a terceira tentativa é feita após backoff exponencial com jitter (base 100 ms, teto 1 s)  
**And** se ainda falhar, o erro final volta ao handler e a mensagem permanece na fila.

### US5 — Falha rápida com circuit breaker

**Given** mais de cinco execuções consecutivas do breaker falharam  
**When** outra mensagem chega dentro dos 15 segundos seguintes  
**Then** o circuito está `OPEN` e `Execute` retorna `gobreaker.ErrOpenState` sem abrir HTTP.

### US6 — Idempotência no redelivery

**Given** a SQS entrega a mesma mensagem mais de uma vez  
**When** o worker envia o evento  
**Then** o header `Idempotency-Key` contém o `correlation_id` (o mesmo valor de `x-cid`)  
**And** a API downstream é responsável por deduplicar por essa chave; o worker não mantém estado local.

### US7 — Encaminhar para DLQ

**Given** `BALANCE_QUEUE_DLQ` configurada e `ApproximateReceiveCount >= 10`  
**When** a mensagem é recebida  
**Then** o core-sdk envia o corpo para a DLQ e deleta a mensagem da fila original, sem executar o handler.

## Requisitos funcionais

- **RF-01** Criar o consumidor com `sqs.NewConsumer(ctx, config.New(), con.handle)`, injetando o cliente API no handler.
- **RF-02** Validar `processing_code` e `amount[].description` antes de desreferenciá-los, retornando `ErrInvalidMessage`.
- **RF-03** Converter `TransactionCreated` em `ledger.Event` preenchendo `Cid`, `OrgID`, `ProgramID`, `AccountID`, `ProcessingCode`, `Producer`, `Amounts` e `Fees`.
- **RF-04** Enviar `EventRequest` com `processing_code`, `program_id`, `account_id`, `producer`, `amounts` e `fees`.
- **RF-05** Definir os headers `Content-Type: application/json`, `x-cid`, `x-tenant` e `Idempotency-Key`.
- **RF-06** Aceitar somente `201 Created` como sucesso.
- **RF-07** Retentar até 3 vezes somente em `429`, `5xx` e erros de rede, com backoff exponencial e jitter.
- **RF-08** Proteger a chamada HTTP com circuit breaker nomeado `LedgerEventsAPI`, compartilhado entre mensagens.
- **RF-09** Não deletar mensagens cujo handler retornou erro; mover para DLQ após 10 recebimentos quando configurada.
- **RF-10** Emitir spans e logs estruturados com identificadores da mensagem e do evento.
- **RF-11** Fechar o cliente HTTP da API no graceful shutdown.

## Entidades e contratos

### Entrada — `TransactionCreated` (`internal/infra/message/types.go`)

| Campo JSON | Tipo | Destino no `ledger.Event` |
|---|---|---|
| `correlation_id` | `string` | `Cid` |
| `org_id` | `string` | `OrgID` |
| `program.id` | `int64` | `ProgramID` |
| `account_id` | `int64` | `AccountID` |
| `processing_code` | `*string` | `ProcessingCode` (validado antes do uso) |
| `amount[].description` | `*string` | chave de `Amounts` (validada antes do uso) |
| `amount[].value` | `json.Number` | valor de `Amounts` via `decimal.NewFromString` |
| `tax[].type` | `string` | chave de `Fees` |
| `tax[].value` | `json.Number` | valor de `Fees` via `decimal.NewFromString` |

`Producer` é fixo: `"regular"`. Outros campos do payload são decodificados, mas não usados na conversão. Valores ausentes ou inválidos retornam `ErrInvalidMessage`.

### Domínio — `ledger.Event`

`Cid`, `OrgID`, `ProgramID`, `AccountID`, `ProcessingCode`, `Producer`, `Amounts map[string]decimal.Decimal`, `Fees map[string]decimal.Decimal`.

### Saída — `EventRequest` (`POST /v1/ledger/events`)

`processing_code`, `program_id`, `account_id`, `producer`, `amounts`, `fees`. `Cid` e `OrgID` **não** vão no corpo; seguem nos headers `x-cid` e `x-tenant`, com `Idempotency-Key` repetindo o correlation id. O corpo é serializado com `encoding/json`.

## Requisitos não funcionais

- **RNF-01** Valores monetários são `decimal.Decimal` no domínio e na requisição, sem passar por `float64`.
- **RNF-02** Três spans OTel encadeados: `Message::Handler` → `Service::Handler` → `LedgerEventsApi::CreateEvent`.
- **RNF-03** Cliente HTTP instrumentado pelo core-sdk (`httpclient.Client()`).
- **RNF-04** Circuit breaker: `ConsecutiveFailures > 5`, `Timeout = 15s`, transições logadas; o estado sobrevive entre mensagens.
- **RNF-05** Retry usa `context.Context` para abortar o backoff durante shutdown.
- **RNF-06** Processamento concorrente e DLQ são fornecidos pelo core-sdk; este repositório fornece a configuração.

## Critérios de aceite

- `go test ./...` passa sem serviços externos.
- `go test ./internal/infra/message` cobre conversão válida, precisão decimal e mensagens inválidas.
- `go test ./internal/infra/ledger/events` cobre `201`, `400` sem retry, retry `5xx → 201`, circuit breaker aberto e envio de `Idempotency-Key`.
- Nenhum teste depende de Docker, LocalStack ou mockserver.

## Limitações conhecidas

- A deduplicação é contratual: o worker envia `Idempotency-Key`, mas não verifica duplicatas localmente. Se a API downstream não deduplicar, redelivery gera novos eventos.
- O retry em processo ocupa o worker por no máximo ~300 ms por mensagem; falhas finais voltam para a SQS.
- O atributo de span `event_type` recebe a struct `ledger.Event` inteira.
- Não há métricas específicas para retry, circuit breaker ou mensagens inválidas.
