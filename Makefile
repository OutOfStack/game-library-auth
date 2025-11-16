### Database
dockerrunpg:
	docker-compose up -d db

### Auth service
build:
	mkdir -p bin
	go build -o bin/game-library-auth cmd/game-library-auth/main.go

build-mng:
	mkdir -p bin
	go build -o bin/game-library-auth-manage cmd/game-library-auth-manage/main.go

run:
	go run ./cmd/game-library-auth/.

dockerbuildauth:
	docker build -t game-library-auth:latest .

dockerrunauth:
	docker compose up -d auth

test:
	go test -v -race ./...

cover:
	go test -cover -coverpkg=./... -coverprofile=coverage.out ./... && go tool cover -func=coverage.out

SWAG_VERSION := v1.16
SWAG_PKG := github.com/swaggo/swag/cmd/swag@$(SWAG_VERSION)
generate-swag:
	@swag --version >/dev/null 2>&1 || { echo "Installing swag..."; go install ${SWAG_PKG}; }
	@echo "Found swag, generating documentation..."
	swag init \
	-d cmd/game-library-auth,internal/handlers,internal/web

MOCKGEN_VERSION := v0.6
MOCKGEN_PKG := go.uber.org/mock/mockgen@$(MOCKGEN_VERSION)
generate-mocks:
	@mockgen -version >/dev/null 2>&1 || { echo "Installing mockgen..."; go install ${MOCKGEN_PKG}; }
	@echo "Found mockgen, generating mocks..."
	mockgen -source=internal/handlers/auth.go -destination=internal/handlers/mocks/auth.go -package=handlers_mocks
	mockgen -source=internal/handlers/unsubscribe.go -destination=internal/handlers/mocks/unsubscribe.go -package=handlers_mocks
	mockgen -source=internal/facade/provider.go -destination=internal/facade/mocks/provider.go -package=facade_mocks
	mockgen -source=pkg/database/tx.go -destination=pkg/database/mocks/tx.go -package=database_mocks

generate: generate-swag generate-mocks

LINT_VERSION := v2.6
LINT_PKG := github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(LINT_VERSION)
lint:
	@golangci-lint version >/dev/null 2>&1 || { echo "Installing golangci-lint..."; go install ${LINT_PKG}; }
	@echo "Found golangci-lint, running..."
	golangci-lint run

### Manage service
# apply all migrations
migrate:
	go run ./cmd/game-library-auth-manage/. -from-file migrate

# rollback last migration
rollback:
	go run ./cmd/game-library-auth-manage/. -from-file rollback

keygen:
	go run ./cmd/game-library-auth-manage/. keygen

secretgen:
	go run ./cmd/game-library-auth-manage/. secretgen

dbuildmng:
	docker build -f Dockerfile.mng -t game-library-auth-mng:latest .
