# Spec 002 — API de health (liveness/readiness)

**Status:** Implementado (reverse-engineered)  
**Fonte:** `internal/infra/rest/server/server.go`, `cmd/main.go`, `scripts/k8s/deployment.yaml`, `core-sdk v0.8.8`  
**Commit de referência:** `5794727`

## Visão geral

O worker expõe um servidor Echo com **apenas dois endpoints de health**. Não existem rotas de negócio HTTP: toda a funcionalidade chega pela fila SQS (spec 001). O servidor escuta em `APP_PORT` (padrão `5002`) e é desligado no graceful shutdown com contexto de 25 s. O verificador de readiness é o `message.Consumer` criado no startup e injetado em `server.New`, não um consumidor novo por requisição.

## Cenários de uso

### US1 — Liveness

**Given** o processo está em execução  
**When** `GET /health/live` é chamado  
**Then** a resposta é `200` sem corpo, sem consultar dependências externas.

### US2 — Readiness com SQS disponível

**Given** a fila `BALANCE_QUEUE` responde a `GetQueueAttributes`  
**When** `GET /health/ready` é chamado  
**Then** a resposta é `200` sem corpo.

### US3 — Readiness com SQS indisponível

**Given** `GetQueueAttributes` falha ou excede 500 ms  
**When** `GET /health/ready` é chamado  
**Then** a resposta é `503` sem corpo.

## Requisitos funcionais

- **RF-01** Registrar `GET /health/live` e `GET /health/ready` no `echo.Echo`.
- **RF-02** Liveness retorna `200` incondicionalmente.
- **RF-03** Readiness verifica a saúde da fila via `HealthChecker` injetado, implementado por `message.Consumer`.
- **RF-04** Limitar a verificação de readiness a 500 ms com `context.WithTimeout` e `select` no resultado.
- **RF-05** Retornar `503` quando a verificação falhar.
- **RF-06** Aplicar os interceptadores de tracing e logging do core-sdk a todas as requisições.
- **RF-07** Configurar um `CustomValidator` que combine `go-playground/validator` com a interface `shared.EntityRequest`.
- **RF-08** Desligar o servidor no graceful shutdown usando o contexto de 25 s de `cmd/main.go`.

## Critérios de aceite

- `GET /health/live` responde `200` mesmo sem LocalStack em execução.
- `GET /health/ready` responde `503` quando o endpoint AWS configurado está fora do ar.
- Os probes do Kubernetes em `scripts/k8s/deployment.yaml` usam `/health/live` e `/health/ready` na porta nomeada `ledger-worker` (`5002`).
- Não há outra rota registrada em `server.routes()`.
- Readiness não cria novos clientes AWS por requisição.

## Limitações conhecidas

- O timeout de 500 ms é curto para uma chamada AWS real e é fixo no código.
- A verificação roda em goroutine; se `Health()` travar, a requisição retorna `503` no timeout, mas a goroutine pode ficar pendurada.
- `CustomValidator` está registrado, mas não há rotas de negócio nem tipos que implementem `EntityRequest` hoje.
- Não há endpoint de métricas HTTP; métricas são exportadas via OTLP/gRPC pelo core-sdk (spec 003).
- Liveness não detecta falha do consumidor SQS: se o consumidor parar, o processo ainda responde `200`.
