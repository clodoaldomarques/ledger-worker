worker = ledger-worker
repository = clodoaldomarques

up:
	docker compose up -d
	$(MAKE) terraform

down: 
	docker compose down -v

run:
	export $$(cat .env | xargs) && go run cmd/main.go

build:
	docker build -t $(repository)/$(worker):$(version) -f scripts/docker/worker/Dockerfile .
	docker tag $(repository)/$(worker):$(version) $(repository)/$(worker):latest

push:
	docker push $(repository)/$(worker):$(version)
	docker push $(repository)/$(worker):latest

publish: build push

version:
	docker images | grep $(worker)

restart: down up

logs:
	docker compose logs $(container)

terraform:
	@if [ ! -d "scripts/terraform/.terraform" ]; then \
		echo "▶️  Inicializando Terraform..."; \
		terraform -chdir=scripts/terraform/ init; \
	else \
		echo "✅ Terraform já inicializado (pulando init)."; \
	fi
	@echo "⏳ Aguardando LocalStack na porta 4566..."
	@until nc -z localhost 4566; do echo "⏳ esperando..."; sleep 2; done
	@echo "📋 Gerando plano..."
	terraform -chdir=scripts/terraform/ plan
	@echo "🚀 Aplicando..."
	terraform -chdir=scripts/terraform/ apply -auto-approve

test:
	go test ./... -coverprofile cover.out
	go tool cover -html=cover.out