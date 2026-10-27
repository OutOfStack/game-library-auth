# game-library-auth

[![Build](https://img.shields.io/github/actions/workflow/status/OutOfStack/game-library-auth/main.yml?branch=main&label=build)](https://github.com/OutOfStack/game-library-auth/actions/workflows/main.yml)
[![codecov](https://codecov.io/gh/OutOfStack/game-library-auth/graph/badge.svg?token=NRIKDH6OFR)](https://codecov.io/gh/OutOfStack/game-library-auth)

## Introduction

`game-library-auth` is an authentication service for the `game-library` web application. It is responsible for user authentication and authorization.

`game-library` consists of:
- [game-library](https://github.com/OutOfStack/game-library) - main service for fetching, storing, and providing games data
- current service handles authentication and authorization
- [game-library-ui](https://github.com/OutOfStack/game-library-ui) - UI service
- [game-library-infra](https://github.com/OutOfStack/game-library-infra) - infrastructure and deployment configurations


## Table of Contents

- [Introduction](#introduction)
- [Installation](#installation)
- [Usage](#usage)
- [Tech Stack](#tech-stack)
- [Configuration](#configuration)
- [Documentation](#documentation)
- [Contributing](#contributing)
- [List of Make Commands](#list-of-make-commands)
- [License](#license)

## Installation

Prerequisites:  [`go`](https://go.dev/doc/install), [`Docker`](https://docs.docker.com/desktop/), [`Make`](https://www.gnu.org/software/make/). 

To set up the service, follow these steps:

1. Clone the repository:
   ```bash
   git clone https://github.com/OutOfStack/game-library-auth.git
   cd game-library-auth
   ```

2. Set up the database:
   ```bash
   make drunpg # runs postgres server with 'auth' db in docker container
   make migrate # applies all migrations to database
   ```

3. Generate key pair for local JWT signing:
   ```bash
   make keygen # creates private/public key pair files
   ```

4. Create the `.env` file based on [`.example.env`](./.example.env) and update it with your local configuration settings.

5. Get Google API Client ID for Google OAuth and set it in `.env`:
   https://developers.google.com/identity/gsi/web/guides/get-google-api-clientid

6. Get GitHub OAuth App Client ID and Secret for GitHub OAuth and set them in `.env`:
   https://docs.github.com/en/apps/oauth-apps/building-oauth-apps/creating-an-oauth-app

7. Get Resend API key and set it along with the sender details in `.env`:
    https://resend.com/api-keys

8. Build and run the service:
   ```bash
   make build
   make run
   ```
   Notification emails are delivered by a separate worker:
   ```bash
   make run-worker
   ```
   `make run-worker` overrides `APP_DEBUG_ADDRESS` to `localhost:6062` for the worker's debug server, so it can run alongside the API.

Refer to the [List of Make Commands](#list-of-make-commands) for a complete list of commands.

## Usage

After installation, you can use the following Make commands to develop the service:

- `make test`: Runs tests for the whole project.
- `make generate`: Generates proto files, documentation for Swagger UI and mocks for testing.
- `make lint`: Runs golangci-lint for code analysis.

Refer to the [List of Make Commands](#list-of-make-commands) for a complete list of commands.

## Tech Stack and Integrations

- Data storage with PostgreSQL.
- Transactional email delivery through Resend API.
- gRPC for internal service-to-service communication.
- Tracing with OTLP exporter (Jaeger).
- Log management with Graylog.
- Metrics collection with Prometheus.
- Code analysis with golangci-lint.
- CI/CD with GitHub Actions and deploy to Kubernetes (microk8s) cluster.

## Configuration

- The service can be configured using `.env` or environment variables, described in [`settings.go`](./internal/appconf/settings.go)
- CI/CD configs are in [`./github/workflows/`](./.github/workflows/)
- k8s deployment configs are in [`./k8s`](./.k8s/)
- k8s secrets in [`secrets.yaml`](./.k8s/secrets.yaml) are not applied by CI. Create them in the cluster before deploying

Both the API and the worker serve operational endpoints on a separate debug server (`APP_DEBUG_ADDRESS`, port 6060 in k8s):
`/metrics` (Prometheus), `/liveness` and `/readiness` (k8s probes), `/debug/vars` and `/debug/pprof`. Readiness fails until all listeners are bound
and again as soon as shutdown starts. Outside the cluster the debug server is reachable only through the basic-auth ingress (`/_debug/auth/`).
The `game-library-infra` Prometheus discovers pods in the `game-library` namespace by `prometheus.io/scrape`, `prometheus.io/port` and `prometheus.io/path`
annotations, scraping only the declared TCP container port equal to the annotated one. The pod `app` label becomes the job name (`game-library-auth` or `game-library-auth-worker`).

## Documentation

API documentation is available via [Swagger UI](http://localhost:8001/swagger/index.html). To generate the documentation, run:
```bash
make generate
```

### gRPC

The service exposes a gRPC endpoint for internal service-to-service communication.

**Endpoint:** `localhost:9001` (`APP_GRPC_ADDRESS` environment variable in [`.example.env`](./.example.env))

**Protobuf Definition:** [`api/proto/authapi/v1/authapi.proto`](./api/proto/authapi/v1/authapi.proto)

**Testing with grpcurl:**

```bash
# list services
grpcurl -plaintext localhost:9001 list

# describe service
grpcurl -plaintext localhost:9001 describe authapi.v1.AuthApiService

# call VerifyToken method
grpcurl -plaintext -d '{"token": "your-jwt-token-here"}' -emit-defaults localhost:9001 authapi.v1.AuthApiService/VerifyToken
```

## Contributing

Commit message and release rules are documented in [CONTRIBUTING.md](./CONTRIBUTING.md).

## List of Make Commands

#### Main Commands
    build      builds app
    build-mng  builds manage app
    run        runs app
    test       runs tests for the whole project
    generate   generates docs for Swagger UI
    lint       runs golangci-lint
    cover      outputs tests coverage

#### Database Commands
    drunpg     runs postgres server with 'auth' db in docker container
    migrate    applies all migrations to database (reads from config file)
    rollback   roll backs one last migration of database (reads from config file)

#### Key Management
    keygen     creates private/public key pair files
    secretgen  generates a cryptographically secure random secret for HMAC

#### Docker Commands
    dbuildauth builds auth app docker image
    dbuildmng  builds manage app docker image
    drunauth   runs auth app in docker container

## License

[MIT License](./LICENSE.md)
