# Plano técnico — Spec 002

> Este plano documenta a implementação **atual**, não trabalho futuro.

## Contexto técnico

- Servidor: `echo/v4 v4.16.0`, inicializado em `server.New()` com `s.routes()`.
- Porta: `APP_PORT`, padrão `5002`; iniciada em goroutine separada em `cmd/main.go`.
- Dependências: `core-sdk/pkg/otel/tracer`, `core-sdk/pkg/zap/logger`, `core-sdk/pkg/aws/sqs`, `go-playground/validator`.
- Deploy: Kubernetes com `livenessProbe` e `readinessProbe` via HTTP.

## Decisões técnicas atuais

| Área | Decisão | Local |
|---|---|---|
| Rotas | Somente health; sem grupo de rotas de negócio | `internal/infra/rest/server/server.go` |
| Códigos | `NoContent(http.StatusOK)` no sucesso; `NoContent(http.StatusServiceUnavailable)` na falha | `server.go` |
| Checagem SQS | `GetQueueAttributes` com atributo `QueueArn`, via core-sdk | `core-sdk/pkg/aws/sqs/consumer.go` |
| Timeout | `context.WithTimeout` de 500 ms por chamada de readiness | `server.go` |
| Interceptadores | `tracer.Interceptor()` e `logger.Interceptor()` registrados globalmente | `server.go` |
| Validação | `CustomValidator` + `shared.EntityRequest`, sem uso efetivo hoje | `server.go`, `rest/shared` |
| Shutdown | `echo.Echo.Shutdown(ctx)` com 25 s | `cmd/main.go` |

## Fluxo de requisição

```text
GET /health/live  → interceptadores → Liveness  → 200
GET /health/ready → interceptadores → Readiness
                      → ctx com 500 ms
                      → message.NewConsumer(ctx).Health()
                      → GetQueueAttributes(QueueArn)
                      → 200 (ok) | 503 (erro/timeout)
```

## Configuração relevante

| Variável | Uso |
|---|---|
| `APP_PORT` | Porta do Echo; padrão `5002`. |
| `BALANCE_QUEUE` | URL usada pelo consumidor criado na readiness. |
| `AWS_ADDRESS`, `AWS_REGION`, `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY` | Cliente SQS usado na readiness. |

## Verificação

- Sem testes automatizados específicos para o servidor REST neste commit.
- Verificação manual: subir o worker com `.env` e chamar `curl -i http://localhost:5002/health/live` e `/health/ready`.
- Com LocalStack fora: liveness deve continuar `200`, readiness deve retornar `503`.

## Riscos e dívida técnica

1. Criação de consumidor AWS por requisição de readiness.
2. Timeout de 500 ms possivelmente insuficiente.
3. Liveness independente do estado real do consumidor SQS.
4. Código de validação sem uso efetivo.

## Cobertura dos requisitos

| Requisito | Evidência |
|---|---|
| RF-01 a RF-03, RF-06 a RF-07 | `internal/infra/rest/server/server.go` |
| RF-04, RF-05 | `Readiness` |
| RF-08 | `cmd/main.go` |
