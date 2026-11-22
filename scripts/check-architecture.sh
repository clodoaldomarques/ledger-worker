#!/usr/bin/env bash
# Guarda arquitetural da spec 005: o core não pode depender de adapters nem de
# bibliotecas de transporte/configuração. Imprime arquivo e linha infratores.
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repo_root"

forbidden=(
	"github.com/clodoaldomarques/ledger-worker/internal/adapters"
	"github.com/clodoaldomarques/ledger-worker/config"
	"github.com/aws/"
	"github.com/labstack/echo"
	"github.com/clodoaldomarques/core-sdk/pkg/aws"
	"github.com/clodoaldomarques/core-sdk/pkg/otel/echo"
	"github.com/clodoaldomarques/core-sdk/pkg/otel/httpclient"
	"github.com/clodoaldomarques/core-sdk/pkg/env"
)

status=0

if [[ -d internal/core ]]; then
	while IFS= read -r file; do
		for pattern in "${forbidden[@]}"; do
			if matches="$(grep -nF "$pattern" "$file")"; then
				printf 'arch-check: import proibido %s em %s\n' "$pattern" "$file" >&2
				printf '%s\n' "$matches" >&2
				status=1
			fi
		done
	done < <(find internal/core -type f -name '*.go' | sort)
fi

if [[ "$status" -ne 0 ]]; then
	exit 1
fi

echo "arch-check OK: internal/core sem dependências de adapters, SDK de nuvem, Echo, config ou transporte"
