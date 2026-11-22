# Tarefas — Spec 003

**Legenda:** `[x]` verificado no código; `[ ]` lacuna ou trabalho pendente.

## Validação da spec

- [x] Confirmar variáveis, defaults e singleton em `config/config.go`.
- [x] Confirmar que `make run` carrega `.env` e que `go run` não carrega.
- [x] Confirmar todos os alvos de teste, benchmark e profiling do `Makefile`.
- [x] Confirmar imagem, runtime e `EXPOSE 5002` no Dockerfile.
- [x] Confirmar os sete serviços do Docker Compose e o scrape do Prometheus.
- [x] Confirmar recursos e endpoints do Terraform.
- [x] Confirmar Deployment, Service e ConfigMap em `scripts/k8s/`.
- [x] Confirmar as três origens divergentes do endpoint do LocalStack.
- [x] Confirmar que o mockserver não cobre `POST /v1/ledger/events`.
- [x] Executar `go test ./...` sem serviços externos.

## Lacunas conhecidas

- [ ] Unificar o endpoint do LocalStack em `.env`, Compose, Terraform, `make terraform` e `make send-event`.
- [ ] Adicionar expectation de `POST /v1/ledger/events` em `expectations.json`.
- [ ] Criar `.env.example` com as variáveis documentadas.
- [ ] Separar geração de cobertura da abertura do HTML em `make test` (ou criar alvo específico para CI).
- [ ] Preencher ou remover `scripts/k8s/configmap.yaml`.
- [ ] Padronizar `OTEL_SERVICE_NAME` entre local (`ledger-events`) e Kubernetes (`ledger-worker`).
- [ ] Expor `net/http/pprof` ou corrigir `make profile-goroutine`, que hoje aponta para uma porta inexistente.
- [ ] Adicionar CI com `go test ./...` e verificação de formatação.
- [ ] Documentar/localizar o estado do Terraform para evitar dependência de máquina específica.
