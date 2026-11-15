# Repository Guidelines

## Project Structure & Module Organization

Read `README.md` for service context. This repository contains the Go auth service of game library; UI, the main game library API and infrastructure usually live in parent folder under `game-library*` repos.

- `cmd/api/`, `cmd/worker/` and `cmd/manage/`: API service (HTTP + gRPC), notification worker and database-management entry points.
- `internal/api/`: HTTP handlers (`auth/`, `unsubscribe/`), gRPC server (`grpc/`) and health probes (`tools/`); `internal/api/openapi/`: code generated from the HTTP spec.
- `internal/facade/`: business logic; `internal/database/`: PostgreSQL persistence; `internal/model/`: domain models; `internal/auth/`: JWT and unsubscribe tokens.
- `internal/client/`, `internal/worker/` and `internal/appconf/`: external integrations (Google, GitHub, Resend, InfoAPI), outbox email worker and configuration.
- `internal/web/`: shared HTTP response, validation types and embedded HTML templates.
- `pkg/`: shared utilities and generated protobuf code; `api/proto/`: protobuf sources; `api/http/openapi.yaml`: HTTP API spec (source of truth).
- `scripts/migrations/`: SQL migrations; `examples/`: manual `.http` requests.
- `.github/workflows/` and `.k8s/`: CI and deployment configuration.

## Build, Test, and Development Commands

Use Go 1.27, Docker and Make. Copy `.example.env` to `.env` and configure local connections before starting.

- `make dockerrunpg`: start local PostgreSQL; `make migrate` / `make rollback`: apply or roll back migrations.
- `make keygen`: create JWT signing key pair files; `make secretgen`: print a random HMAC secret (e.g. for `EMAIL_SENDER_UNSUBSCRIBE_SECRET`).
- `make build` / `make build-worker` / `make build-mng`: compile binaries into `bin/`.
- `make run` / `make run-worker`: start the API or worker from source.
- `make test`: run all tests with the race detector; `make cover`: print coverage.
- `make lint`: run golangci-lint using `.golangci.yaml`; for gci and format errors use `goimports`.
- `make generate`: regenerate OpenAPI server code (`make generate-openapi`), mocks (`make generate-mocks`) and protobufs (`make generate-proto`).

## Coding Style & Naming Conventions

Use standard Go naming, tab indentation and configured `gofmt`, `goimports` and `gci` formatters. Write comments for all exported functions and structs. Start comments inside functions with a lowercase letter.

Change the HTTP API spec-first: edit `api/http/openapi.yaml` (document operations, request bodies and all responses), run `make generate-openapi`, then implement the generated `ServerInterface` methods on `auth.API` or `unsubscribe.API`. Request/response models are generated from the spec; declare request validation as go-playground/validator tags via `x-oapi-codegen-extra-tags` and mirror constraints in the schema.

Never hand-edit generated files (`*.gen.go`, `*.pb.go`, mocks). After protobuf changes run `make generate-proto`, which uses `buf.yaml` and `buf.gen.yaml`. Server protobuf definitions live in `api/proto/authapi/v*/`; client code for external services (InfoAPI) is generated from their remote repositories pinned in `Makefile`; generated code goes to `pkg/proto/*api/v*/`. gRPC is served on `APP_GRPC_ADDRESS` (9001 locally). For new mocks, add a line to the `generate-mocks` target in `Makefile`. Run `go mod tidy` after adding or removing dependencies.

## Testing Guidelines

Use Go's `testing`, Testify (`assert`/`require`) and GoMock. Write tests for all exported functions, including error paths. Place `*_test.go` files beside the source and use external test packages (`<package>_test`). Use `t.Context()` and `t.Parallel()`.

Repository tests in `internal/database` require Docker and start PostgreSQL with testcontainers.

## Validation Before Finishing

Run and fix issues found by:

- `make build`
- `go test -v -race ./... | grep -E "(FAIL:|RUN.*failed|panic:|error:)"` (use `go test -v ./...` if `-race` fails)
- `make lint`
- `make generate | grep -E "(error:|warning:|failed)"` when the OpenAPI spec, protobufs or mocked interfaces changed

Update `README.md` when changes significantly affect what it describes.

## Commit & Pull Request Guidelines

Follow `CONTRIBUTING.md` and Conventional Commits: `fix:`, `feat:`, `docs:`, `chore:`, `ci:`. Breaking changes use `!` or `BREAKING CHANGE:` and affect release versions.

Name branches `<type>/<short-kebab-description>`, where `<type>` is the Conventional Commit type of the main change (`feat`, `fix`, `chore`, `ci`, `docs`, `refactor`, `test`), e.g. `feat/openapi-spec-first`, `fix/unsubscribe-templates`. Branch off `main`; keep the description to a few words naming the change, not the ticket or tool.

## Configuration

Configuration is loaded from `.env` or environment variables via `appconf.Load`; each binary loads its own struct (`appconf.API`, `Worker`, `Manage`). Every field of these structs is a section type with a value-receiver `Validate` method; add a new setting to an existing section or create a section with `Validate`.

To introduce a config param, add it to `.example.env`, then to `internal/appconf/settings.go`, then to `.k8s/config.yaml`. Secrets go to `.k8s/secrets.yaml` instead (not applied by CI); a key reaches a container only if its Secret is listed in that container's `envFrom` or the key is referenced via `secretKeyRef` in `.k8s/deployment.yaml` (API and worker load different Secrets). Keep credentials out of commits.

## Workflow Restrictions

- Do not run modifying `git` commands (`git add`, `git commit`, `git push`, etc.) without user request.
- Do not delete files; notify the user when files become redundant.

## File Review Boundaries

- Do not review or analyze generated code matching `**/*.gen.go` or `**/*.pb.go`, mocks (`**/mocks/**`, `**/*_mock.go`) or vendored dependencies (`vendor/**`).
- Do not read sensitive files: `*.pem`, `*.key`, `.env`, `app.env`.
