# Especificação da Funcionalidade: Adoção da Arquitetura Hexagonal — `ledger-worker`

**Branch da funcionalidade**: `005-hexagonal-adapters`

**Criado**: 2026-11-22

**Repo**: `github.com/clodoaldomarques/ledger-worker`

**Status**: Proposta — especificação **prescritiva** (estado desejado), ainda não implementada

**Entrada**: `docs/specs/` (specs `001`-`003`), `AGENTS.md` equivalente e engenharia reversa deste repo.

> **Escopo**: esta é a spec 005 aplicada a `ledger-worker`. Os IDs `US1`-`US5`, `FR-001`..`FR-014` e `SC-001`..`SC-007` são idênticos nos repos da família para rastreabilidade; o conteúdo abaixo descreve **somente este repo**. Nenhuma mudança de comportamento externo está autorizada.

## Contexto e Motivação

`ledger-worker` já possui composition root em `cmd/main.go` e uma porta de saída (`EventsProvider`, em `internal/application/ledger/interfaces.go`), mas a topologia ainda é `domain`/`application`/`infra`: o consumer SQS vive em `infra/message`, o cliente da API ledger em `infra/ledger/events`, e o servidor de health em `infra/rest/server`. O consumer depende de `*service.Service` em vez de uma porta de entrada, e `server.New(con)` recebe o adapter concreto.

Isso dificulta trocar o cliente HTTP da API ledger ou o consumer SQS por um fake, e duplicar a convenção dos demais repos da família. A migração é estrutural: mover para `internal/core/ledger/{domain,application,ports}` e `internal/adapters/{inbound,outbound}`, explicitar `ProcessEventUseCase` e `HealthChecker`, e ajustar a raiz de composição que já existe. O circuit breaker, retries, contrato de eventos e endpoint de health permanecem intactos.

## Cenários de Usuário e Testes *(obrigatório)*

### US1 — Trocar adaptador de saída sem alterar casos de uso (Prioridade: P1)

Como mantenedor de `ledger-worker`, quero substituir o cliente HTTP da API ledger (ou o publisher SQS) por outro adaptador ou fake sem modificar `service.go`.

**Por que esta prioridade**: `EventsProvider` já é uma porta, mas seu nome não expressa intenção e o core ainda convive com pacotes de transporte; tornar o contrato explícito é a garantia central da arquitetura hexagonal.

**Teste independente**: executar a mesma suíte de `service_test.go` com um fake de `LedgerEventPublisher` e com o adapter `ledger-api` real; ambos devem satisfazer `CreateEvent(ctx, ledger.Event)`.

**Cenários de aceitação**:

1. **Dado** `internal/core/ledger/**`, **Quando** inspecionar seus imports, **Então** não há import de `internal/adapters`, AWS SDK, Echo, `config` nem `core-sdk` de transporte.
2. **Dado** um fake que implementa `LedgerEventPublisher`, **Quando** injetado no `Service`, **Então** o processamento de eventos se comporta como hoje.
3. **Dado** o adapter `ledger-api` real, **Quando** compilado, **Então** satisfaz integralmente a porta, mantendo circuit breaker e retries.
4. **Dado** um método novo na porta, **Quando** algum adapter não o implementar, **Então** o compilador falha — sem fallback dinâmico.

---

### US2 — Adicionar adaptador de entrada sem duplicar regra (Prioridade: P1)

Como mantenedor, quero que o consumer SQS e um futuro adaptador de entrada chamem o mesmo caso de uso pela porta `ProcessEventUseCase`.

**Por que esta prioridade**: hoje o consumer conhece `*service.Service`; sem porta de entrada, um segundo adaptador poderia replicar validação e sequencing de eventos.

**Teste independente**: um adaptador de entrada alternativo invoca `ProcessEventUseCase` e verifica o mesmo resultado do consumer real, com o mesmo `Service`.

**Cenários de aceitação**:

1. **Dado** o consumer SQS e um segundo adaptador, **Quando** ambos chamam o caso de uso, **Então** utilizam a mesma implementação e as mesmas regras.
2. **Dado** um DTO de qualquer adaptador de entrada, **Quando** convertido para `ledger.Event`, **Então** a validação permanece no domínio.
3. **Dado** um teste do adaptador SQS, **Quando** ele injeta um fake de `ProcessEventUseCase`, **Então** não são necessários AWS ou LocalStack.

---

### US3 — Compor dependências uma única vez na raiz (Prioridade: P1)

Como operador/mantenedor, quero que `cmd/main.go` continue sendo a única raiz de composição, agora com adapters injetados em vez de construídos por fluxo.

**Por que esta prioridade**: a raiz já existe (`message.NewConsumer(ctx)` e `server.New(con)`), mas ainda chama construtores que escondem configuração e clientes internamente; controlar ciclo de vida e shutdown exige composição explícita.

**Teste independente**: subir o worker com adapters fake e verificar uma instância de cada porta por processo e fechamento explícito no graceful shutdown.

**Cenários de aceitação**:

1. **Dado** `cmd/main.go`, **Quando** a aplicação inicia, **Então** constrói config, `LedgerEventsApi`, publisher, `Service`, consumer SQS e servidor, injetando as dependências.
2. **Dado** o consumer, **Quando** uma mensagem chega, **Então** ele não chama construtores de adapter; recebe `ProcessEventUseCase` por parâmetro.
3. **Dado** o endpoint de health, **Quando** chamado, **Então** consulta `HealthChecker` injetado em vez do adapter concreto.
4. **Dado** o sinal de shutdown, **Quando** o graceful shutdown termina, **Então** o `Close()` do consumer e do servidor é chamado pela raiz.

---

### US4 — Preservar contrato externo e levantar guarda arquitetural (Prioridade: P2)

Como revisor, quero garantir que a migração não altere o contrato de processamento de eventos nem o de health, e que a regra de dependência continue válida.

**Por que esta prioridade**: a mudança é estrutural; retries, circuit breaker e contrato de mensagem são comportamento observável.

**Teste independente**: `go test ./...`, `go build ./...`, `go vet ./...`, `make arch-check` e execução da suíte de eventos das specs `001`-`003`.

**Cenários de aceitação**:

1. **Dado** o contrato de evento atual, **Quando** a migração termina, **Então** campos, ordenação e comportamento de retry/circuit breaker permanecem idênticos.
2. **Dado** o contrato de publicação via `CreateEvent`, **Quando** comparado ao estado atual, **Então** não há mudança de payload ou tratamento de erro.
3. **Dado** um import de `core` para `adapters`, **Quando** `make arch-check` roda, **Então** o comando falha com arquivo e linha infratores.
4. **Dado** o `git diff`, **Quando** revisado, **Então** `internal/core/ledger/domain` contém apenas movimento de pacote/imports, sem mudança de regra.

---

### US5 — Remover código morto e dependências órfãs (Prioridade: P2)

Como mantenedor, quero que nenhuma porta ou dependência sem uso sobreviva à migração.

**Por que esta prioridade**: este repo não tem `gringotts`, `ProgramAPI` nem `ollama`, mas a renomeação de `EventsProvider` pode deixar aliases, mocks ou imports órfãos.

**Teste independente**: `go mod tidy` e busca por `gringotts`, `Griphook`, `ProgramAPI`, `ollama` e `EventsProvider` não retornam ocorrências antigas não intencionais.

**Cenários de aceitação**:

1. **Dado** `EventsProvider`, **Quando** a renomeação para `LedgerEventPublisher` termina, **Então** nenhuma referência antiga permanece em código, mocks ou documentação.
2. **Dado** `mock.go`, **Quando** regenerado, **Então** contém apenas mocks das portas efetivamente usadas.
3. **Dado** `go.mod`, **Quando** `go mod tidy` roda, **Então** dependências não usadas desaparecem.
4. **Dada** alguma funcionalidade do fluxo de eventos, **Quando** comparada ao estado anterior, **Então** não houve alteração.

---

### Casos Limite

- O consumer e o cliente HTTP passam a ser compartilhados por toda a vida do processo; confirmar thread safety das clientes AWS do `core-sdk` antes da Fase 3.
- `config.New()` continua `sync.Once`; mudar esse comportamento está fora do escopo.
- `tracer`/`logger` permanecem dentro do core nesta spec; decorators são fase opcional futura.
- O circuit breaker e os retries de `infra/ledger/events` devem continuar no adaptador de saída, não no core.
- Bugs já documentados nas specs `001`-`003` não são corrigidos aqui.
- `go generate` e `mock.go` seguem a diretiva em `interfaces.go`; o arquivo gerado nunca é editado à mão.

## Requisitos *(obrigatório)*

### Requisitos Funcionais

- **FR-001**: O core (`internal/core/ledger/**`) NÃO DEVE importar `internal/adapters`, AWS SDK, Echo, `config` nem pacotes de transporte do `core-sdk`.
- **FR-002**: As portas DEVEM ser declaradas em `internal/core/ledger/ports` e implementadas pelos adapters; a dependência aponta sempre do adapter para o core.
- **FR-003**: O caso de uso DEVE ser exposto por `ProcessEventUseCase` no core, consumido pelo adaptador SQS e por qualquer outro adaptador de entrada.
- **FR-004**: A raiz de composição (`cmd/main.go`) DEVE instanciar cada adapter uma única vez e injetá-lo no `Service`, no consumer e no servidor.
- **FR-005**: O consumer e o servidor NÃO DEVEM construir adapters nem casos de uso; recebem as portas por parâmetro/construtor.
- **FR-006**: Health checks DEVEM ser expostos como porta `HealthChecker` injetada e consumida pelo adaptador HTTP.
- **FR-007**: O mapeamento de erros de domínio para resposta HTTP DEVE ficar no adaptador de entrada.
- **FR-008**: As portas DEVEM ter nomes que expressem intenção: `LedgerEventPublisher`, `ProcessEventUseCase`, `HealthChecker` — em vez de `EventsProvider` genérico.
- **FR-009**: Este repo não contém `ProgramAPI`/`gringotts`/`ollama`; nenhuma porta órfã DEVE ser introduzida e `mock.go` DEVE ser regenerado somente a partir das portas em uso.
- **FR-010**: O comportamento observável NÃO PODE mudar: contrato de eventos, retries, circuit breaker, payload de `CreateEvent` e respostas de health.
- **FR-011**: DEVE existir `make arch-check` neste repo, executado antes do merge.
- **FR-012**: `docs/specs/README.md`, `platform.md` e os planos das specs `001`-`003` afetados DEVEM ser atualizados na mesma mudança de código.
- **FR-013**: A limpeza de portas/mocks NÃO DEVE alterar funcionalidade observável no processamento de eventos.
- **FR-014**: Este repo DEVE manter um único `spec.md`, `plan.md` e `tasks.md` coerentes com o layout alvo, mesmo que a implementação ocorra em PRs separados.

### Entidades Principais

- **Porta de entrada (driving port)**: `ProcessEventUseCase` no core.
- **Porta de saída (driven port)**: `LedgerEventPublisher`, `HealthChecker` no core.
- **Adapter de entrada**: consumer SQS (`adapters/inbound/sqs`) e servidor HTTP de health (`adapters/inbound/http/server`).
- **Adapter de saída**: `adapters/outbound/ledger-api`, com circuit breaker e retries inalterados.
- **Composition root**: `cmd/main.go`; único lugar que conhece adapters concretos e configuração.
- **Guarda de arquitetura**: `scripts/check-architecture.sh`, exposta por `make arch-check`.

## Critérios de Sucesso *(obrigatório)*

### Resultados Mensuráveis

- **SC-001**: `go test ./...`, `go build ./...` e `go vet ./...` passam neste repo antes e depois da migração.
- **SC-002**: Zero imports de `internal/adapters` a partir de `internal/core` neste repo.
- **SC-003**: Zero chamadas a `NewConsumer`, `events.New` ou outros construtores de adapter fora de `cmd/**` e dos próprios adapters.
- **SC-004**: Diff de `internal/core/ledger/domain` restrito a movimento de pacote/imports — nenhuma regra alterada.
- **SC-005**: Nenhuma diferença no contrato de processamento de eventos e nas respostas de health.
- **SC-006**: `go mod tidy` sem dependências órfãs e sem ocorrências de `gringotts`/`ProgramAPI`/`Griphook`/`ollama` neste repo.
- **SC-007**: `make arch-check` falha quando um import proibido é introduzido artificialmente e passa no estado final.

## Suposições

- Os clientes AWS do `core-sdk` e da AWS SDK v2 são seguros para reutilização concorrente; confirmar na Fase 3.
- O contrato externo atual é o correto, inclusive desvios já registrados; correções viram specs separadas.
- Este repo é um módulo Go independente (`go.mod` próprio).
- Toda a documentação necessária para este repo está neste diretório; não há dependência de links externos.
- Não há CI; a guarda arquitetural é executada por comando local e pelo revisor.
