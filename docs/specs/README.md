# Specs — ledger-worker

Este diretório contém especificações no formato do GitHub Spec Kit (`spec.md`, `plan.md`, `tasks.md`). As specs `001`-`003` foram reconstruídas da implementação original; a spec `005` foi escrita como estado desejado e logo implementada — todas descrevem o código atual.

> Commit de referência das specs `001`-`003`: `5794727`. A spec `005` foi implementada na branch `005-hexagonal-adapters` (2026-11-22).

## Convenção

| Arquivo | Significado neste projeto |
|---|---|
| `spec.md` | Comportamento observável, requisitos e critérios de aceite extraídos do código, config e manifests. |
| `plan.md` | Arquitetura e decisões técnicas adotadas, documentadas como estão. |
| `tasks.md` | Checklist de validação da spec contra o código. `[x]` = verificado; `[ ]` = lacuna ou trabalho pendente. |

Como são specs de software já implementado, o status é `Implementado (reverse-engineered)`. Quando o README sugeria algo que o código não faz, o texto diz explicitamente.

A verificação local inclui `go test ./...`, `go build ./...`, `go vet ./...` e `make arch-check` (spec 005). A guarda arquitetural falha com arquivo e linha quando `internal/core/**` importa adapters, SDK de nuvem, Echo, `config` ou transporte do `core-sdk`. `make test` não é adequado para automação porque abre o relatório HTML no navegador.

## Índice

| ID | Capacidade | Status |
|---|---|---|
| [001](001-processamento-de-eventos/) | Processamento de eventos SQS → Ledger Events API | Implementado, com limitações conhecidas |
| [002](002-api-de-health/) | API de health (liveness/readiness) | Implementado |
| [003](003-configuracao-e-infraestrutura/) | Configuração, infraestrutura local e deploy | Implementado, com divergências de endpoint |
| [005](005-hexagonal-adapters/) | Migração para arquitetura hexagonal (bundle local com os IDs da família) | Implementada |

A spec 005 foi a única prescritiva: descreveu o estado desejado e foi concluída na branch `005-hexagonal-adapters`. A topologia adotada (`core` + `adapters` + portas) está resumida em [`platform.md`](platform.md).

Princípios que atravessam as capacidades: [`constitution.md`](constitution.md).
