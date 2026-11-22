# Tarefas: `ledger-worker` — Adoção da Arquitetura Hexagonal

**Repo**: `github.com/clodoaldomarques/ledger-worker` | **Ordem**: 2 de 4

**Branch**: `005-hexagonal-adapters`

**Entrada**: [spec.md](spec.md) e [plan.md](plan.md) deste diretório; `docs/specs/` deste repo.

**Nota**: este repo já possui composition root em `cmd/main.go` e porta consumidora `EventsProvider`; o trabalho é alinhar topologia, nomes e dono da composição.

**Convenção**: `[ ]` = a fazer; `[x]` = concluído **no código**. Cada checkpoint exige `go test ./...`, `go build ./...` e `go vet ./...` verdes.

**IDs**: `T001`-`T031` são os marcos comuns da família, adaptados ao escopo deste repo; `LW-001`-`LW-021` são os passos concretos que os implementam aqui.

## Fase 0: Baseline e guarda

- [ ] T001 Rodar `go test ./...`, `go build ./...` e `go vet ./...` neste repo e registrar o resultado antes da mudança
- [ ] T002 [P] Criar `scripts/check-architecture.sh`: falha se `internal/core/**` importar `internal/adapters`, SDK de cloud, Echo ou `config`
- [ ] T003 [P] Adicionar alvo `arch-check` ao `Makefile` e documentá-lo em `docs/specs/`
- [ ] T004 Confirmar thread safety dos clientes AWS compartilhados entre o consumer e o servidor
- [ ] LW-001 Rodar `go test ./...`, `go build ./...`, `go vet ./...` e registrar o estado antes da mudança
- [ ] LW-002 Criar `scripts/check-architecture.sh` e o alvo `make arch-check` (core não importa adapters)

**Checkpoint**: suíte verde e guarda arquitetural passando no estado atual.

---

## Fase 1: Movimento físico

- [ ] T005 [P] Executar o mapeamento de pastas de [plan.md](plan.md) neste repo, sem alterar regras de negócio
- [ ] T006 [P] Ajustar todos os imports e a diretiva `go:generate` (que vai junto com `interfaces.go` para `ports/`)
- [ ] T007 Rodar `go generate ./...`, `go test ./...` e `go build ./...` após o movimento
- [ ] LW-003 `git mv internal/domain/ledger → internal/core/ledger/domain`
- [ ] LW-004 `git mv internal/application/ledger/service.go → internal/core/ledger/application/service.go`
- [ ] LW-005 `git mv internal/application/ledger/interfaces.go → internal/core/ledger/ports/outbound.go` (com `go:generate`); criar `inbound.go` com `ProcessEventUseCase`
- [ ] LW-006 `git mv internal/infra/message/consumer.go → internal/adapters/inbound/sqs/consumer.go`; mover `errors.go`/`types.go` e testes junto
- [ ] LW-007 `git mv internal/infra/ledger/events → internal/adapters/outbound/ledger-api` (mantendo circuit breaker e retries)
- [ ] LW-008 `git mv internal/infra/rest/server → internal/adapters/inbound/http/server`; mover `rest/shared` para `adapters/inbound/http/shared`
- [ ] LW-009 Mover mocks para `internal/core/ledger/ports/mock.go` e ajustar imports em `service_test.go`/`bench_test.go`

**Não aplicáveis aqui**: T008 (reparo do build de `balances-worker`); T020-T023 (remoção de `gringotts`/`ProgramAPI`/`ollama`, ausentes neste repo).

**Checkpoint**: suíte verde; nenhuma mudança de comportamento no consumer.

---

## Fase 2: Portas explícitas

- [ ] T009 [P] Renomear portas de saída para nomes intencionais (`LedgerEventPublisher`, `HealthChecker`)
- [ ] T010 [P] Declarar a porta de entrada `ProcessEventUseCase` no core
- [ ] T011 [P] Mudar construtores dos adapters para receber cliente/config em vez de chamar `config.New()` internamente
- [ ] T012 [P] Regenerar mocks ao lado das portas e ajustar `service_test.go`/`bench_test.go`
- [ ] T013 [US1] Criar teste de contrato da porta com fake in-memory executando a mesma suíte do caso de uso
- [ ] LW-010 Renomear `EventsProvider` → `LedgerEventPublisher` na porta de saída
- [ ] LW-011 Declarar `HealthChecker` no core; implementá-lo no consumer SQS e no servidor HTTP
- [ ] LW-012 Declarar `ProcessEventUseCase` no core e garantir que o consumer dependa da porta, não de `*service.Service`
- [ ] LW-013 Trocar construtores dos adapters para receber cliente/config; `events.New(ctx)` e `sqs.NewConsumer` passam a ser chamados pela raiz
- [ ] LW-014 Regenerar mocks e adicionar teste de contrato com fake da porta de saída

**Checkpoint**: core sem dependência de `core-sdk` de transporte, SQS ou HTTP.

---

## Fase 3: Composition root e adaptadores de entrada

- [ ] T014 [P] Refinar a construção de adapters em `cmd/main.go` (raiz já existente neste repo)
- [ ] T015 [US2][P] Fazer o consumer depender da porta de entrada em vez da implementação concreta
- [ ] T016 [US3][P] Injetar `HealthChecker` no servidor HTTP de health
- [ ] T017 [US3] Centralizar o mapeamento de erros no adaptador de entrada
- [ ] T018 [US3] Garantir que `Close()` do consumer e do servidor seja chamado no graceful shutdown
- [ ] T019 [US2] Adicionar teste do adaptador SQS com fake da porta de entrada
- [ ] LW-015 Em `cmd/main.go`, construir config, `LedgerEventsApi`, publisher, `Service`, consumer SQS e servidor; injetar dependências
- [ ] LW-016 Mover `NewConsumer` da lógica de composição para `cmd/main.go`, mantendo o circuit breaker compartilhado
- [ ] LW-017 Ajustar `server.New(con)` para receber `HealthChecker` em vez do adapter concreto
- [ ] LW-018 Garantir `Close()` do consumer no graceful shutdown
- [ ] LW-019 Adicionar teste do adaptador SQS com fake de `ProcessEventUseCase`

**Checkpoint**: SC-003; contrato de processamento de eventos inalterado.

---

## Fase 4: Opcional / diferido (fora dos critérios de sucesso)

- [ ] T024 [P] Avaliar decorators para mover `tracer`/`logger` do core para as bordas
- [ ] T025 [P] Avaliar divisão de portas grossas por capacidade
- [ ] T026 Corrigir bugs conhecidos das specs `001`-`003` em specs próprias, com coordenação de contrato

---

## Fase 5: Documentação e verificação

- [ ] T027 [P] Registrar a regra de dependência via portas na documentação de arquitetura deste repo e atualizar versão/data
- [ ] T028 [P] Atualizar `docs/specs/README.md`, `platform.md` e os planos das specs `001`-`003` com os novos caminhos
- [ ] T029 [P] Registrar a convenção arquitetural neste repo
- [ ] T030 Rodar `go test ./...`, `go build ./...`, `go vet ./...`, `make arch-check`
- [ ] T031 Conferir o contrato de processamento de eventos e as respostas de health; validar SC-001 a SC-007 aplicáveis
- [ ] LW-020 Atualizar os planos das specs `001`-`003` com os novos caminhos e registrar a convenção hexagonal em `docs/specs/README.md`
- [ ] LW-021 Rodar `go test ./...`, `go build ./...`, `go vet ./...`, `make arch-check`

## Dependências e Ordem de Execução

- Fase 0 bloqueia todas as demais.
- Fase 1 deve estar verde antes da Fase 2.
- Fase 2 bloqueia Fase 3; o graceful shutdown existente deve ser preservado em todas as fases.
- `ledger-config` é executado antes deste repo; o padrão é replicado aqui.
- Fase 4 depende de decisão explícita e não é necessária para fechar a spec.
