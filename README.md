# Ledger Worker

> Event-driven Go worker for asynchronous ledger processing, integrating Amazon SQS, backend services, distributed tracing and resilient HTTP communication.

## Overview

Ledger Worker is a Go-based backend service responsible for asynchronously processing ledger-related events received through Amazon SQS.

The service consumes messages from an SQS queue, validates and transforms the incoming event into a domain representation, and forwards it to the Ledger Events API.

The project was created to explore practical patterns commonly used in distributed backend systems, including:

- Event-driven architecture
- Asynchronous message processing
- Microservices integration
- Domain-oriented application structure
- Circuit breaker
- Distributed tracing
- Structured logging
- Graceful shutdown
- Infrastructure as Code
- Kubernetes deployment
- Local AWS-compatible development with LocalStack

## Architecture

The worker follows a layered structure separating domain logic from infrastructure concerns.

```text
                         ┌──────────────────────┐
                         │    Balance Service   │
                         └──────────┬───────────┘
                                    │
                                    │ Event
                                    ▼
                         ┌──────────────────────┐
                         │      Amazon SQS      │
                         │   balance-sqs-queue  │
                         └──────────┬───────────┘
                                    │
                              Message Consumer
                                    │
                                    ▼
                         ┌──────────────────────┐
                         │    Ledger Worker     │
                         │                      │
                         │       Go Service     │
                         └──────────┬───────────┘
                                    │
                          Deserialize / Transform
                                    │
                                    ▼
                         ┌──────────────────────┐
                         │    Domain Service    │
                         │                      │
                         │   Ledger Event       │
                         └──────────┬───────────┘
                                    │
                             HTTP + Circuit Breaker
                                    │
                                    ▼
                         ┌──────────────────────┐
                         │    Ledger Events     │
                         │         API          │
                         └──────────────────────┘
```

The worker also integrates with OpenTelemetry for distributed observability.

## Event Processing Flow

The main processing pipeline is:

```text
SQS Message
    │
    ▼
Message Handler
    │
    ▼
Deserialize Event
    │
    ▼
Build Domain Event
    │
    ├── Correlation ID
    ├── Organization ID
    ├── Program ID
    ├── Account ID
    ├── Processing Code
    ├── Amounts
    └── Fees
    │
    ▼
Ledger Domain Service
    │
    ▼
Ledger Events API
    │
    ▼
HTTP Response
```

The incoming message is transformed into a domain-oriented `ledger.Event` before being sent to the downstream service.

Amounts and fees are represented using decimal arithmetic to avoid relying on floating-point values for financial calculations.

## Domain Model

The domain layer contains the core representation of a ledger event:

```go
type Event struct {
    Cid            string
    OrgID          string
    ProgramID      int64
    AccountID      int64
    ProcessingCode string
    Producer       string
    Amounts        map[string]decimal.Decimal
    Fees           map[string]decimal.Decimal
}
```

The domain is intentionally isolated from infrastructure implementations.

This allows the application logic to depend on abstractions rather than directly coupling business behavior to HTTP or messaging implementations.

## Project Structure

```text
ledger-worker/
├── cmd/
│   └── main.go
│
├── config/
│
├── internal/
│   ├── domain/
│   │   └── ledger/
│   │       ├── event.go
│   │       ├── interface.go
│   │       └── service.go
│   │
│   └── infra/
│       ├── ledger/
│       │   └── events/
│       ├── message/
│       └── rest/
│
├── docs/
│   └── model/
│
├── scripts/
│   ├── docker/
│   ├── k8s/
│   └── terraform/
│
├── docker-compose.yaml
├── Makefile
├── go.mod
└── go.sum
```

### `cmd/`

Application entry point and service initialization.

The application starts the HTTP server, initializes OpenTelemetry and starts the SQS consumer.

### `internal/domain/`

Contains domain-oriented components and business abstractions.

The domain layer does not depend directly on the messaging or HTTP infrastructure.

### `internal/infra/`

Contains infrastructure implementations such as:

- SQS message handling
- Ledger Events API integration
- REST server
- External service communication

### `scripts/`

Contains infrastructure and local development resources:

- Docker configuration
- Kubernetes manifests
- Terraform configuration

## Resilience

One of the main goals of the project is to explore resilience patterns in distributed systems.

### Circuit Breaker

Communication with the Ledger Events API is protected by a circuit breaker using `sony/gobreaker`.

The circuit transitions after consecutive failures and temporarily prevents additional requests from reaching an unhealthy downstream service.

```text
                  ┌─────────────────┐
                  │  Ledger Worker  │
                  └────────┬────────┘
                           │
                           ▼
                   ┌───────────────┐
                   │ Circuit       │
                   │ Breaker       │
                   └───────┬───────┘
                           │
                    ┌──────┴──────┐
                    │             │
                 CLOSED          OPEN
                    │             │
                    ▼             ▼
              HTTP Request      Fail Fast
                    │
                    ▼
             Ledger Events API
```

The current configuration trips the circuit after more than five consecutive failures and keeps it open for a configured timeout.

## Asynchronous Processing

The worker consumes events asynchronously through Amazon SQS.

The SQS abstraction is provided by the `core-sdk` shared library.

This allows the worker to focus on domain processing while keeping messaging infrastructure encapsulated.

The architecture supports the separation:

```text
Messaging Infrastructure
          │
          ▼
    Message Handler
          │
          ▼
     Domain Service
          │
          ▼
 External Integration
```

This separation makes the application easier to evolve and test.

## Observability

The application integrates with OpenTelemetry for distributed tracing.

Tracing is propagated through the message processing flow and the downstream HTTP request.

Relevant contextual information such as correlation identifiers and tenant information is associated with the processing flow.

Structured logging is provided through the shared `core-sdk` library.

The local development environment includes:

- OpenTelemetry Collector
- Prometheus
- Zipkin

This allows the complete processing flow to be observed locally.

## Graceful Shutdown

The worker handles operating system termination signals and performs a graceful shutdown.

The shutdown flow includes:

1. Receive `SIGINT` or `SIGTERM`
2. Stop accepting new HTTP requests
3. Allow active operations to finish
4. Shutdown OpenTelemetry
5. Release application resources
6. Complete service shutdown

The application currently provides a configurable timeout for the graceful shutdown process.

## Infrastructure

The project includes infrastructure definitions for both local development and Kubernetes environments.

### LocalStack

LocalStack provides AWS-compatible local services for development and testing.

The Docker Compose environment enables:

- Amazon SQS
- Amazon SNS
- DynamoDB

The worker communicates with LocalStack through a configurable AWS endpoint.

### Terraform

Terraform is used to provision the SQS queues used by the application.

Current infrastructure includes:

- `balance-sqs-queue`
- `balance-sqs-dlq-queue`

The project uses the HashiCorp AWS provider configured to work with LocalStack.

### Kubernetes

The project includes Kubernetes manifests for:

- Deployment
- Service
- ConfigMap

The worker deployment is configured with:

- Container resource requests and limits
- AWS configuration
- SQS queue configuration
- Ledger Events API endpoint
- OpenTelemetry endpoint
- Service name

## Local Development

### Requirements

- Go 1.26+
- Docker
- Docker Compose
- Terraform
- AWS CLI
- Kubernetes / Minikube (optional)

### Clone

```bash
git clone https://github.com/clodoaldomarques/ledger-worker.git
cd ledger-worker
```

### Start the local infrastructure

```bash
make up
```

This starts the Docker Compose development environment and provisions the required Terraform resources.

### Run the application

Configure the required environment variables and run:

```bash
make run
```

Alternatively:

```bash
go run cmd/main.go
```

### Run tests

```bash
make test
```

Or directly:

```bash
go test ./...
```

### Build the Docker image

```bash
make build
```

### Deploy to Kubernetes

```bash
make apply
```

### Remove the Kubernetes deployment

```bash
make destroy
```

## Sending a Test Event

The project provides a Make target for sending an example event to the SQS queue:

```bash
make send-event
```

This allows the complete asynchronous processing flow to be exercised locally.

## Technology Stack

| Technology | Purpose |
|---|---|
| Go | Backend service |
| Amazon SQS | Asynchronous messaging |
| LocalStack | Local AWS environment |
| Terraform | Infrastructure as Code |
| Kubernetes | Container orchestration |
| Docker | Containerization |
| OpenTelemetry | Distributed tracing |
| Prometheus | Metrics |
| Zipkin | Trace visualization |
| Echo | HTTP server |
| GoBreaker | Circuit breaker |
| shopspring/decimal | Financial decimal arithmetic |
| Testify | Testing |
| Core SDK | Shared infrastructure components |

## Engineering Concepts

This project demonstrates practical implementation of:

- Event-driven architecture
- Asynchronous processing
- Microservices communication
- Domain-oriented design
- Separation of domain and infrastructure
- Dependency inversion
- Circuit breaker
- Distributed tracing
- Structured logging
- Graceful shutdown
- Infrastructure as Code
- Containerization
- Kubernetes deployment
- Local cloud simulation

## Design Decisions

### Why SQS?

SQS decouples event producers from the worker responsible for processing ledger events.

This provides asynchronous processing and allows producers and consumers to evolve independently.

### Why a Circuit Breaker?

The Ledger Events API is an external dependency from the worker's perspective.

The circuit breaker prevents the worker from continuously sending requests to an unhealthy downstream service and allows the failure to be isolated.

### Why Decimal?

Financial amounts should not rely on binary floating-point arithmetic.

The project uses `shopspring/decimal` to represent amounts and fees with decimal semantics.

### Why OpenTelemetry?

The worker participates in a distributed processing flow.

Tracing provides visibility into the relationship between message consumption, domain processing and downstream API calls.

## Reliability Considerations

Asynchronous systems introduce failure scenarios that need to be explicitly considered.

This project uses an SQS queue and a dedicated DLQ to support failure isolation.

Message processing should be designed with idempotency in mind because distributed messaging systems can deliver a message more than once.

The worker therefore treats message processing and downstream operations as distributed-system concerns rather than assuming exactly-once delivery.

## Future Improvements

Possible improvements for the project include:

- Idempotency handling for message processing
- Automated integration tests using LocalStack
- End-to-end test scenarios
- Health and readiness probes
- Improved Kubernetes deployment strategy
- Horizontal scaling based on queue depth
- Automated CI/CD pipeline
- Enhanced metrics around message processing
- Retry/backoff policies for downstream API failures

## Project Status

This project is part of my Go backend engineering portfolio and serves as a practical exploration of event-driven architecture, asynchronous processing, distributed systems and cloud-native infrastructure.

The project is intentionally designed to demonstrate engineering concepts rather than represent a production-ready financial platform.

## Author

**Clodoaldo Marques**

Backend Software Engineer focused on Go, Microservices, Distributed Systems and Cloud-Native architectures.

- GitHub: https://github.com/clodoaldomarques
- LinkedIn: https://www.linkedin.com/in/clodoaldomarques/