# Constitution — ledger-worker

> Princípios extraídos da implementação atual e revisados pela spec 005 em 2026-11-22. Servem para julgar mudanças futuras e manter as specs coerentes. A topologia atual está em [`platform.md`](platform.md).

## I. Regra de dependência por portas (hexagonal)

`internal/core/ledger/domain` contém apenas o modelo `Event`. A orquestração fica em `internal/core/ledger/application`; as portas em `internal/core/ledger/ports` (`ProcessEventUseCase`, `LedgerEventPublisher`, `HealthChecker`). Adaptadores em `internal/adapters/inbound/**` (SQS e HTTP) e `internal/adapters/outbound/ledger-api` dependem do core — nunca o contrário. `cmd/main.go` é a única raiz de composição e injeta cada adaptador uma vez. O serviço de aplicação depende de `LedgerEventPublisher`, nunca de `LedgerEventsApi`. `make arch-check` aplica a guarda.

## II. Aritmética decimal para valores financeiros

`Amounts` e `Fees` usam `shopspring/decimal`, convertidos a partir de `json.Number` com `decimal.NewFromString`. Não reintroduzir `float64` no caminho entre o JSON e o domínio (ver spec 001).

## III. Observabilidade propagada

Spans OTel nos pontos `Message::Handler`, `Service::Handler` e `LedgerEventsApi::CreateEvent`. O correlation id trafega em `x-cid` e o tenant em `x-tenant`. Logging estruturado é responsabilidade do `core-sdk`.

## IV. Resiliência explícita

O circuit breaker `sony/gobreaker` protege a API downstream: abre após mais de cinco falhas consecutivas e fica aberto por 15 segundos. A DLQ é acionada pelo `core-sdk` quando `ApproximateReceiveCount >= 10` e `BALANCE_QUEUE_DLQ` está configurada.

## V. Configuração por ambiente, com singleton conhecido

`config.New()` usa `sync.Once`. As opções `With...` mutam a instância compartilhada e afetam o processo inteiro. Preferir variáveis de ambiente; em testes, evitar ou isolar essa mutação.

## VI. Verificação sem serviços externos

Os testes usam `httptest` e mocks: `go test ./...` não depende de Docker, LocalStack ou rede. `make test` também abre o relatório HTML no navegador e não deve ser usado em automação.

## VII. Infraestrutura declarada, mas divergente

O endpoint do LocalStack aparece em três lugares diferentes: `.env`/Docker Compose (`localhost:4566`), Terraform e `make terraform` (`192.168.67.2:30002`) e `make send-event` (`192.168.49.2:30002`). Toda mudança de endpoint deve atualizar os três.

## Governança

- Mudanças no fluxo de eventos exigem atualização da spec 001.
- Mudanças em rotas HTTP ou probes exigem atualização da spec 002.
- Mudanças em config, Makefile, Compose, Terraform ou Kubernetes exigem atualização da spec 003 e do princípio VII.
- Mudanças nos princípios I–IV exigem revisão explícita deste documento antes do código.
