# Tarefas — Spec 002

**Legenda:** `[x]` verificado no código; `[ ]` lacuna ou trabalho pendente.

## Validação da spec

- [x] Confirmar que apenas `/health/live` e `/health/ready` estão registrados.
- [x] Confirmar que liveness retorna `200` sem depender de SQS.
- [x] Confirmar que readiness usa o `HealthChecker` injetado, com `GetQueueAttributes(QueueArn)` e timeout de 500 ms.
- [x] Confirmar os códigos `200` e `503`.
- [x] Confirmar interceptadores de tracing e logging.
- [x] Confirmar que o shutdown do Echo usa o contexto de 25 s.
- [x] Confirmar que os probes do Kubernetes apontam para essas rotas na porta `5002`.

## Lacuna corrigida

- [x] Reutilizar um único consumidor/cliente SQS para readiness em vez de criar um por requisição (injeção em `server.New`).

## Lacunas restantes

- [ ] Reavaliar o timeout de 500 ms e permitir configuração.
- [ ] Fazer a liveness refletir a saúde do consumidor SQS, se desejado.
- [x] Adicionar testes automatizados para liveness e readiness (`server_test.go`, com fake de `ports.HealthChecker`).
- [ ] Remover ou efetivamente usar `CustomValidator`/`EntityRequest`.
