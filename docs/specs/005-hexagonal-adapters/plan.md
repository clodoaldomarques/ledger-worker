# Plano de Implementação: Adoção da Arquitetura Hexagonal — `ledger-worker`

**Branch**: `005-hexagonal-adapters` | **Data**: 2026-11-22 | **Repo**: `github.com/clodoaldomarques/ledger-worker`

**Entrada**: [spec.md](spec.md) deste diretório, specs `001`-`003` de `docs/specs/` e a estrutura atual deste repo.

**Nota**: este plano descreve o trabalho **executado** neste repo; o progresso real está marcado em `tasks.md`. Todo o conteúdo necessário está aqui e nos arquivos vizinhos deste diretório. A Fase 4 (decorators/divisão de portas/correções de bugs) permanece diferida e fora dos critérios de sucesso.

## Resumo

Alinhar `ledger-worker` à topologia hexagonal: `internal/core/ledger` (domínio + aplicação + portas), `internal/adapters` (consumer SQS e health como entrada; cliente da API ledger como saída) e `cmd/main.go` como composition root — que já existe e deve ser refinado, não recriado. `EventsProvider` vira `LedgerEventPublisher`, `ProcessEventUseCase` e `HealthChecker` entram no core. Circuit breaker, retries, payload de eventos e health permanecem congelados.

## Contexto Técnico

**Linguagem/Versão**: Go (ver `go.mod` deste repo)

**Dependências principais**: AWS SDK v2 (SQS), `core-sdk` (tracer/logger/métricas/HTTP), Echo v4 (health), `go.uber.org/mock` (mockgen), OpenTelemetry

**Armazenamento**: nenhum banco local; o worker publica eventos na API ledger e consome fila SQS via LocalStack em desenvolvimento

**Testes**: `go test ./...`, `go build ./...`, `go vet ./...`; suíte determinística com mocks gerados, sem Docker/AWS obrigatório

**Processamento**: consumer SQS em goroutine, servidor HTTP de health em outra, graceful shutdown com timeout (25s) e `con.Close()` — comportamento a preservar

**Plataforma alvo**: worker em container

**Restrições**: sem CI, `config.New()` é singleton `sync.Once`, circuit breaker/retries ficam no adaptador de saída, tracing dentro do core permanece nesta fase.

**Escopo neste repo**: 1 movimento de pacotes, 3 portas, 2 adaptadores de entrada, 1 adaptador de saída, 1 guarda de arquitetura.

## Verificação da Constituição

*GATE: aplicado antes da Fase 0 e reavaliado ao fim da Fase 6.*

| Princípio | Situação planejada | Evidência alvo neste repo |
|---|---|---|
| I. Separação em camadas | **Reescrito** para "regra de dependência via portas" | `internal/core/ledger/**` sem imports de adapters/SQS/Echo/config; raiz em `cmd/main.go` |
| II. O domínio valida, a interface transporta | Passa | DTOs e conversão de evento ficam em `adapters/inbound/sqs`; validação em `domain` |
| III. Configuração por ambiente, com singleton explícito | Passa (inalterado) | `config.New()` continua `sync.Once`; adapters recebem config na construção |
| IV. Observabilidade é parte do request | Passa (com desvio documentado) | Spans permanecem em casos de uso; health continua no adaptador HTTP |
| V. Consultas são desenhadas no modelo de dados | Passa | Contrato de evento e chamadas à API ledger intactos |
| VI. Testes unitários determinísticos | Passa | Mocks ao lado das portas; fake de `ProcessEventUseCase` no teste do consumer |
| VII. Infraestrutura como código | Passa | Compose/Terraform/k8s sem mudança; `make arch-check` adicionado |

**Ação obrigatória na Fase 6**: registrar a regra de dependência via portas na documentação de arquitetura deste repo e atualizar sua data/versão.

## Estrutura do Projeto

### Documentação (esta funcionalidade, neste repo)

```text
docs/specs/005-hexagonal-adapters/
├── spec.md      # US1-US5, FR-001..FR-014, SC-001..SC-007 aplicados a ledger-worker
├── plan.md      # este documento: layout, mapeamento, riscos, verificação
└── tasks.md     # fases 0-6 com IDs T001-T031 e LW-001-LW-021
```

### Código-fonte alvo

```text
cmd/main.go                              # composition root (já existente)
config/config.go                         # mantido
internal/
├── core/ledger/
│   ├── domain/                          # event.go, erros e regras
│   ├── application/                     # service.go, service_test.go, bench_test.go
│   └── ports/
│       ├── inbound.go                   # ProcessEventUseCase
│       ├── outbound.go                  # LedgerEventPublisher, HealthChecker (+ go:generate)
│       └── mock.go                      # gerado
└── adapters/
    ├── inbound/
    │   ├── sqs/                         # consumer.go, errors.go, types.go, consumer_test.go
    │   └── http/
    │       ├── server/                  # Echo, /health/ready, graceful shutdown
    │       └── shared/                  # utilitários HTTP comuns
    └── outbound/ledger-api/             # cliente HTTP, circuit breaker, retries
scripts/check-architecture.sh            # guarda de imports
```

**Decisão de estrutura**: consumer SQS é adaptador de **entrada** (ele dirige o caso de uso); o cliente da API ledger é adaptador de **saída** (implementa portas chamadas pelo core).

## Mapeamento de Pastas

| Origem | Destino | Observações |
|---|---|---|
| `internal/domain/ledger` | `internal/core/ledger/domain` | Somente `git mv` + imports |
| `internal/application/ledger/service.go` | `internal/core/ledger/application/service.go` | Mover `service_test.go` e `bench_test.go` junto |
| `internal/application/ledger/interfaces.go` | `internal/core/ledger/ports/outbound.go` | `go:generate` vai junto; criar `inbound.go` com `ProcessEventUseCase` |
| `internal/application/ledger/mock.go` | `internal/core/ledger/ports/mock.go` | Regenerar a partir das novas portas |
| `internal/infra/message/consumer.go` | `internal/adapters/inbound/sqs/consumer.go` | Mover `errors.go`, `types.go` e `consumer_test.go` junto |
| `internal/infra/ledger/events` | `internal/adapters/outbound/ledger-api` | Manter circuit breaker e retries |
| `internal/infra/rest/server` | `internal/adapters/inbound/http/server` | `server.New()` passa a receber `HealthChecker` |
| `internal/infra/rest/shared` | `internal/adapters/inbound/http/shared` | Sem mudança de comportamento |

## Riscos e Decisões

| Risco / desvio | Por que existe | Decisão |
|---|---|---|
| `tracer`/`logger` dentro do core | Já presentes; mover exigiria reescrever spans e testes | Manter e registrar como desvio |
| `config.New()` singleton | Comportamento atual | Manter; parar de chamá-lo dentro dos adapters |
| Circuit breaker e retries | São parte do contrato de resiliência | Permanecem no adaptador `ledger-api`, jamais no core |
| `server.New(con)` recebe adapter concreto | Acoplamento atual | Passa a receber `HealthChecker` injetado |
| Renomear `EventsProvider` | Nome genérico e não intencional | Renomear para `LedgerEventPublisher`; regenerar mocks e documentação |
| Composição já existe em `cmd/main.go` | Risco de duplicar bootstrap | Refinar a raiz existente; não criar segundo ponto de composição |
| Mover diretórios quebra imports de uma vez | Mudança física | Fase 1 exclusivamente `git mv` + imports, com build verde antes de seguir |
| Bugs já documentados nas specs `001`-`003` | Registrados como comportamento atual | Congelar; cada correção em spec própria |

## Estratégia de Migração

1. **Fase 0 — Baseline**: suíte verde e `make arch-check` criado (a regra atual já é respeitada).
2. **Fase 1 — Movimento físico**: `git mv` + imports + `go generate`; nenhuma mudança de comportamento no consumer.
3. **Fase 2 — Portas**: renomear `EventsProvider`, declarar `ProcessEventUseCase` e `HealthChecker`, adapters recebem cliente/config.
4. **Fase 3 — Composition root**: refinar `cmd/main.go`; consumer depende da porta; health injetado; `Close()` no shutdown.
5. **Fase 4 — Opcional (fora do SC)**: decorators de tracing, divisão de portas, correções de bugs conhecidos.
6. **Fase 5 — Fechamento**: documentação das specs `001`-`003`, verificações finais.

Cada fase terminou com build e testes verdes. Rollback por commit de fase; as fases 0, 1, 2, 3 e 5 estão registradas em commits separados (`git log --oneline`).

## Verificação

```bash
# executar na raiz do repo

go test ./...
go build ./...
go vet ./...
make arch-check

# após o ajuste de imports e da diretiva go:generate
go generate ./internal/core/ledger/...
go test ./internal/core/ledger/... ./internal/adapters/...
go mod tidy

# rastreabilidade do nome antigo
grep -R 'EventsProvider' --include='*.go' --include='*.md' . || true
```

Resultado: comandos verdes, com `go mod tidy` removendo `github.com/golang/mock`. As únicas ocorrências de `EventsProvider` permanecem em texto histórico das próprias specs 001 e 005; nenhuma referência em código, mocks, README ou constitution.

## Dependências e Ordem de Execução

- A spec deve ser aprovada antes da Fase 0.
- `ledger-config` é o repo de referência e deve ser migrado antes deste.
- Fase 0 bloqueia todas as demais; Fase 1 deve estar verde antes da Fase 2.
- Fase 2 bloqueia Fase 3; o graceful shutdown existente deve ser preservado em todas elas.
- Fase 4 depende de decisão explícita e não é necessária para fechar a spec.
