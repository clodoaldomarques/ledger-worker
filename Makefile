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
		terraform -chdir=scripts/terraform/ init;\
	fi
	until nc -z 192.168.49.2 30002; do echo waiting for localstack; sleep 2; done;
	terraform -chdir=scripts/terraform/ plan
	terraform -chdir=scripts/terraform/ apply -auto-approve

test:
	go test ./... -coverprofile cover.out
	go tool cover -html=cover.out


apply: 
	$(MAKE) terraform
	kubectl apply -f scripts/k8s/

destroy:
	kubectl delete -f scripts/k8s/ --ignore-not-found
	terraform -chdir=scripts/terraform/ destroy -auto-approve

reload: destroy apply

send-event:
	@echo "📤 Enviando evento para a fila 'balance-sqs-queue'..."
	aws --endpoint-url=http://192.168.49.2:30002 sqs send-message \
		--queue-url http://192.168.49.2:30002/000000000000/balance-sqs-queue \
		--message-body file://scripts/docker/localstack/event.json
	@echo "✅ Mensagem enviada!"