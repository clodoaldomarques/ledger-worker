# Constitution — ledger-worker

> Princípios extraídos da implementação atual (commit `5794727`). Servem para julgar mudanças futuras e manter as specs coerentes.

## I. Separação entre domínio, aplicação e infraestrutura

`internal/domain/ledger` contém apenas o modelo `Event`. Regras de orquestração ficam em `internal/application/ledger`; adaptadores SQS/HTTP em `internal/infra`. O serviço de aplicação depende da abstração `EventsProvider`, nunca de `LedgerEventsApi`.

## II. Aritmética decimal para valores financeiros

`Amounts` e `Fees` usam `shopspring/decimal`. Não substituir por `float64` no domínio. Observação: a conversão atual usa `decimal.NewFromFloat` a partir de um `float64` do JSON, então a imprecisão da entrada já chega ao decimal (ver spec 001).

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
