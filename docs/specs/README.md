# Specs — engenharia reversa do ledger-worker

Este diretório contém especificações **reconstruídas a partir da implementação atual**, usando o formato do GitHub Spec Kit (`spec.md`, `plan.md`, `tasks.md`). Elas descrevem o que o código faz hoje — não um projeto novo.

> Commit de referência: `5794727` (`adiciona AGENTS.md e atualiza README`).

## Convenção

| Arquivo | Significado neste projeto |
|---|---|
| `spec.md` | Comportamento observável, requisitos e critérios de aceite extraídos do código, config e manifests. |
| `plan.md` | Arquitetura e decisões técnicas **existentes**, documentadas como estão. |
| `tasks.md` | Checklist de validação da spec contra o código. `[x]` = verificado; `[ ]` = lacuna ou trabalho pendente. |

Como são specs de software já implementado, o status é `Implementado (reverse-engineered)`. Quando o README sugeria algo que o código não faz, o texto diz explicitamente.

## Índice

| ID | Capacidade | Status |
|---|---|---|
| [001](001-processamento-de-eventos/) | Processamento de eventos SQS → Ledger Events API | Implementado, com limitações conhecidas |
| [002](002-api-de-health/) | API de health (liveness/readiness) | Implementado |
| [003](003-configuracao-e-infraestrutura/) | Configuração, infraestrutura local e deploy | Implementado, com divergências de endpoint |
| [005](005-hexagonal-adapters/) | Migração para arquitetura hexagonal (bundle local com os IDs da família) | Proposta, prescritiva — não implementada |

A spec 005 é a única prescritiva: descreve o estado desejado, não o código atual.

Princípios que atravessam as três capacidades: [`constitution.md`](constitution.md).
