# Plano técnico — Spec 003

> Este plano documenta a configuração e a infraestrutura **existentes**, não trabalho futuro.

## Contexto técnico

- Go `1.26.1` (`go.mod`); Dockerfile de build usa `golang:1.27.1-alpine`.
- `core-sdk v0.8.8` lê variáveis de ambiente diretamente para AWS e OpenTelemetry.
- Orquestração de tarefas: apenas `Makefile`; não há CI, lint ou pre-commit.
- Infra local: Docker Compose + Terraform com provider AWS apontando para LocalStack.
- Deploy: kubectl aplicando diretório `scripts/k8s/`.

## Decisões técnicas atuais

| Área | Decisão | Local |
|---|---|---|
| Configuração | `sync.Once` + opções mutáveis | `config/config.go` |
| Carregamento de `.env` | `export $(cat .env | xargs)` apenas em `make run` | `Makefile` |
| Fila/DLQ | Filas criadas no Terraform; redrive feito pelo core-sdk em runtime | `scripts/terraform/sqs.tf`, core-sdk |
| Ambiente local | 7 serviços Compose, sem worker containerizado | `docker-compose.yaml` |
| Build | Multi-stage, runtime Alpine, binário estático | `scripts/docker/worker/Dockerfile` |
| Deploy | Imagem `:latest` com `imagePullPolicy: Always`, secret externo | `scripts/k8s/deployment.yaml` |
| Observabilidade | OTLP gRPC `:4317` → coletor → Zipkin/Prometheus | `scripts/docker/otel/config.yaml` |

## Mapa de comandos

```text
make up        → docker compose up -d → make terraform
make terraform → init (se preciso) → nc 192.168.67.2:30002 → plan → apply
make run       → export .env → go run cmd/main.go
make test      → go test ./... -coverprofile cover.out → go tool cover -html
make bench*    → go test -bench=. -run=^$ ./...
make profile-* → pprof / go tool trace
make build     → docker build -t <repo>/ledger-worker:<version> + tag :latest
make push      → docker push <version> + :latest
make publish   → build + push
make apply     → terraform → kubectl apply -f scripts/k8s/
make destroy   → kubectl delete -f scripts/k8s/ → terraform destroy
make reload    → destroy + apply
make send-event → aws sqs send-message em 192.168.49.2:30002
```

## Endpoints do LocalStack por origem

| Origem | Endereço |
|---|---|
| `.env` (`AWS_ADDRESS`, `BALANCE_QUEUE`) | `http://localhost:4566` |
| Docker Compose | publicação na porta `4566` |
| `scripts/terraform/provider.tf` | `http://192.168.67.2:30002` |
| `Makefile:terraform` (espera) | `192.168.67.2:30002` |
| `Makefile:send-event` | `http://192.168.49.2:30002` |
| `scripts/k8s/deployment.yaml` | `http://localstack:4566` |

Qualquer mudança de endpoint precisa atualizar as linhas aplicáveis dessa tabela e a spec 003.

## Verificação

- Testes: `go test ./...` (herméticos, sem serviços).
- Build: `make build version=0.1.0`.
- Ambiente local: `make up`, depois `docker compose logs localstack mockserver`.
- Infra: `terraform -chdir=scripts/terraform plan` após ajustar os endpoints.
- Deploy: `make apply` com `kubectl` configurado e secret `aws-credentials` presente.

## Riscos e dívida técnica

1. Três endpoints diferentes para o mesmo LocalStack.
2. `.env` sem exemplo versionado.
3. `make test` depende de navegador para exibir cobertura.
4. ConfigMap vazio e `OTEL_SERVICE_NAME` divergente entre local e Kubernetes.
5. Ausência de CI e verificações automáticas de lint/format.

## Cobertura dos requisitos

| Requisito | Evidência |
|---|---|
| RF-01 | `config/config.go` |
| RF-02 a RF-05 | `Makefile` |
| RF-06 | `docker-compose.yaml`, `scripts/docker/*` |
| RF-07 | `scripts/terraform/*` |
| RF-08 | `scripts/k8s/*` |
