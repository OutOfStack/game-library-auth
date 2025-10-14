# game-library-auth

## Introduction

game-library-auth is an authentication service for the game-library web application. It is responsible for user authentication and authorization.

This service is part of a game-library web application:
- [game-library](https://github.com/OutOfStack/game-library) - main service for fetching, storing, and providing games data
- current service handles authentication and authorization
- [game-library-ui](https://github.com/OutOfStack/game-library-ui) - UI representation service


## Table of Contents

- [Introduction](#introduction)
- [Installation](#installation)
- [Usage](#usage)
- [Tech Stack](#tech-stack)
- [Configuration](#configuration)
- [Documentation](#documentation)
- [List of Make Commands](#list-of-make-commands)
- [License](#license)

## Installation

Prerequisites: `go`, `Docker`, `Make`. To set up the service, follow these steps:

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

4. Create the `app.env` file based on [app.example.env](./app.example.env) and update it with your local configuration settings.

5. Get Google API Client ID for Google OAuth and set it in [app.env](./app.env):
   https://developers.google.com/identity/gsi/web/guides/get-google-api-clientid

6. Get Resend API key and set it along with the sender details in [app.env](./app.env):
    https://resend.com/api-keys

7. Build and run the service:
   ```bash
   make build
   make run
   ```

Refer to the [List of Make Commands](#list-of-make-commands) for a complete list of commands.

## Usage

After installation, you can use the following Make commands to develop the service:

- `make test`: Runs tests for the whole project.
- `make generate`: Generates documentation for Swagger UI.
- `make lint`: Runs golangci-lint for code analysis.

Refer to the [List of Make Commands](#list-of-make-commands) for a complete list of commands.

## Tech Stack and Integrations

- Data storage with PostgreSQL.
- Tracing with Zipkin.
- Log management with Graylog.
- Transactional email delivery through Resend API.
- Code analysis with golangci-lint.
- CI/CD with GitHub Actions and deploy to Kubernetes (microk8s) cluster.

## Configuration

- The service can be configured using [app.env](./app.env) or environment variables, described in [settings.go](./internal/appconf/settings.go)
- CI/CD configs are in [./github/workflows/](./.github/workflows/)
- k8s deployment configs are in [./k8s](./.k8s/)

## Documentation

API documentation is available via [Swagger UI](http://localhost:8001/swagger/index.html). To generate the documentation, run:
```bash
make generate
```

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

#### Docker Commands
    dbuildauth builds auth app docker image
    dbuildmng  builds manage app docker image
    drunauth   runs auth app in docker container

## License

[MIT License](./LICENSE.md)
