# Spec 003 — Configuração, infraestrutura local e deploy

**Status:** Implementado (reverse-engineered), com divergências de endpoint documentadas  
**Fonte:** `config/config.go`, `.env`, `Makefile`, `docker-compose.yaml`, `scripts/terraform/*`, `scripts/k8s/*`, `scripts/docker/*`  
**Commit de referência:** `5794727`

## Visão geral

A configuração da aplicação vem de variáveis de ambiente lidas pelo `core-sdk` e centralizadas em `config.New()`. O `Makefile` orquestra infraestrutura local (Docker Compose + Terraform), execução, build de imagem, deploy Kubernetes, benchmarks e profiling. Não há CI, lint ou pre-commit configurados no repositório.

## Configuração da aplicação

`config.New()` é um singleton `sync.Once`: na primeira chamada lê o ambiente; as opções `With...` mutam a instância compartilhada, afetando todo o processo.

| Variável | Default | Uso |
|---|---|---|
| `APP_PORT` | `5002` | Porta do servidor Echo (spec 002). |
| `AWS_ADDRESS` | vazio | Endpoint customizado do AWS SDK. |
| `AWS_REGION` | vazio | Região do cliente AWS. |
| `AWS_ACCESS_KEY_ID` | vazio | Credencial local. |
| `AWS_SECRET_ACCESS_KEY` | vazio | Credencial local. |
| `BALANCE_QUEUE` | vazio | URL da fila principal (spec 001). |
| `BALANCE_QUEUE_DLQ` | vazio | URL da DLQ; vazia desativa o movimento (spec 001). |
| `LEDGER_EVENTS_API_URL` | vazio | Base URL da API downstream (spec 001). |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | — | Consumido pelo `core-sdk` via ambiente. |
| `OTEL_SERVICE_NAME` | — | Consumido pelo `core-sdk` via ambiente. |

`MaxReceiveCount()` retorna `10` fixo; não há variável para alterá-lo.

`.env` está no `.gitignore` e não existe arquivo de exemplo. `make run` carrega o arquivo com `export $(cat .env | xargs)`; `go run cmd/main.go` **não** carrega.

## Ambiente local (Docker Compose)

| Serviço | Porta host | Finalidade |
|---|---|---|
| `localstack` | `4566` | SQS, SNS e DynamoDB (`SERVICES=sns,sqs,dynamodb`). |
| `mockserver` | `1080` | Expectativas em `scripts/docker/mockserver/expectations.json`. |
| `otel-collector` | `4317`, `4318`, `8888` | OTLP gRPC/HTTP, métricas internas; exporta para `debug` e Zipkin. |
| `redis` | `6379` | Disponível para uso futuro. |
| `stackport` | `8080` | UI AWS, depende do healthcheck do LocalStack. |
| `zipkin` | `9411` | Visualização de traces. |
| `prometheus` | `9090` | Scrape do coletor em `otel-collector:8888/metrics`. |

O healthcheck do coletor OTLP escuta em `13133`. O pipeline de traces usa `batch` e exporta para `http://zipkin:9411/api/v2/spans`.

## Terraform

- Provider AWS `~> 5.0` com credenciais `test/test`, `us-east-1` e `skip_*` para LocalStack.
- Todos os endpoints apontam para `http://192.168.67.2:30002` (S3 usa `s3.192.168.67.2.localstack.cloud:30002`).
- Recursos declarados: apenas `balance-sqs-queue` e `balance-sqs-dlq-queue`. Não há `redrive_policy` no Terraform: o movimento para DLQ é feito pelo core-sdk via API (spec 001).
- `make terraform` inicializa se necessário, espera `nc -z 192.168.67.2 30002`, roda `plan` e `apply -auto-approve`.

## Kubernetes

- `deployment.yaml`: imagem `clodoaldomarques/ledger-worker:latest` com `imagePullPolicy: Always`, 1 réplica, estratégia `Recreate`, recursos `128Mi/100m` → `256Mi/200m`.
- Credenciais AWS vêm do secret `aws-credentials`. `AWS_ADDRESS`, `BALANCE_QUEUE`, `BALANCE_QUEUE_DLQ` e `LEDGER_EVENTS_API_URL` usam nomes de serviço do cluster; `OTEL_SERVICE_NAME` é `ledger-worker`.
- Probes: `/health/live` e `/health/ready` na porta nomeada `ledger-worker` (`5002`).
- `service.yaml`: `LoadBalancer`, porta `5002`, `nodePort: 31002`.
- `configmap.yaml` existe, mas está vazio.

`make apply` executa Terraform e depois `kubectl apply -f scripts/k8s/`. `make destroy` remove os manifests com `--ignore-not-found` e roda `terraform destroy -auto-approve`. `make reload` é `destroy` seguido de `apply`.

## Imagem Docker

`make build version=<tag>` usa `scripts/docker/worker/Dockerfile`: build em `golang:1.27.1-alpine`, `CGO_ENABLED=0`, binário em `/ledger-worker`, runtime `alpine:3.20` com certificados e timezone, `EXPOSE 5002`. O alvo também cria a tag `:latest`; `make push` envia ambas. Sem `version=`, as tags ficam vazias e `:latest`.

## Expectativas do mockserver

`expectations.json` define apenas `GET /v1/ledger/config/{101..105}/401` e um fallback `404` para `/v1/ledger/config/.*/.*`. O worker chama `POST /v1/ledger/events` (spec 001), para o qual não há expectation.

## Cenários de uso

### US1 — Subir o ambiente local

**Given** Docker, Compose e Terraform instalados e o endereço `192.168.67.2:30002` alcançável  
**When** `make up` é executado  
**Then** os serviços sobem e as duas filas são provisionadas.

### US2 — Enviar evento de teste

**Given** o ambiente local no ar  
**When** `make send-event` é executado  
**Then** o corpo de `scripts/docker/localstack/event.json` é publicado via AWS CLI em `balance-sqs-queue`.  
**And** o processamento ponta a ponta depende do alinhamento dos endpoints e de uma expectation para `POST /v1/ledger/events`.

### US3 — Publicar e deployar

**Given** `version=` informado e registry acessível  
**When** `make publish` e `make apply` são executados  
**Then** as imagens são publicadas e os manifests são aplicados.

## Requisitos funcionais

- **RF-01** Centralizar variáveis de ambiente em `config.New()` com singleton.
- **RF-02** Fornecer `make up`, `down`, `restart`, `logs`, `run`, `terraform`, `apply`, `destroy`, `reload` e `send-event`.
- **RF-03** Fornecer `make build`, `push` e `publish` com tag explícita.
- **RF-04** Fornecer `make test`, `bench`, `bench-count`, `bench-one`, `bench-baseline`, `bench-compare` e `bench-parallel`.
- **RF-05** Fornecer `make profile-cpu`, `profile-mem`, `profile-mem-inuse`, `profile-goroutine`, `profile-trace` e `profile-clean`.
- **RF-06** Provisionar LocalStack, mockserver, OTLP, Zipkin e Prometheus via Compose.
- **RF-07** Provisionar as duas filas SQS via Terraform.
- **RF-08** Definir Deployment, Service e ConfigMap em `scripts/k8s/`.

## Critérios de aceite

- `go test ./...` passa sem Docker, LocalStack ou rede.
- `make bench-one name=BenchmarkCreateEvent_Success` executa um único benchmark.
- `make build version=0.1.0` gera as tags `0.1.0` e `latest`.
- `make apply` aplica os três manifests em `scripts/k8s/`.

## Divergências e limitações conhecidas

| Item | Situação atual |
|---|---|
| Endpoint LocalStack | `.env`/Compose usam `localhost:4566`; Terraform e `make terraform` usam `192.168.67.2:30002`; `make send-event` usa `192.168.49.2:30002`. |
| Mockserver | Não há expectation para `POST /v1/ledger/events`; evento de teste local retorna `404`. |
| `.env` | Gitignored, sem exemplo no repositório. |
| `make test` | Abre `cover.out` no navegador via `go tool cover -html`; inadequado para automação. |
| `make profile-goroutine` | Assume pprof em `localhost:6060`, mas a aplicação não expõe `net/http/pprof`. |
| CI/qualidade | Sem CI, lint, formatter, typecheck separado ou pre-commit. |
| ConfigMap | Arquivo vazio; nenhuma configuração é injetada por ele. |
| `OTEL_SERVICE_NAME` | `.env` usa `ledger-events`; Kubernetes usa `ledger-worker`. |
| Estado do Terraform | `terraform.tfstate` e providers ficam em `scripts/terraform/` (os arquivos de estado são ignorados pelo Git). |
