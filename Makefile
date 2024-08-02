### Database
dockerrunpg:
	docker-compose up -d db

### Auth service
build:
	mkdir -p bin
	go build -o bin/game-library-auth cmd/game-library-auth/main.go

run:
	go run ./cmd/game-library-auth/.

dockerbuildauth:
	docker build -t game-library-auth:latest .

dockerrunauth:
	docker compose up -d auth

test:
	go test -v ./...

LINT_PKG := github.com/golangci/golangci-lint/cmd/golangci-lint@v1.59.1
LINT_BIN := $(shell go env GOPATH)/bin/golangci-lint
lint:
	@if \[ ! -f ${LINT_BIN} \]; then \
		echo "Installing golangci-lint..."; \
    go install ${LINT_PKG}; \
  fi
	@if \[ -f ${LINT_BIN} \]; then \
  	echo "Found golangci-lint at '$(LINT_BIN)', running..."; \
    ${LINT_BIN} run; \
	else \
    echo "golangci-lint not found or the file does not exist"; \
    exit 1; \
  fi

### Manage service
# apply all migrations
migrate:
	go run ./cmd/game-library-auth-manage/. migrate

# rollback last migration
rollback:
	go run ./cmd/game-library-auth-manage/. rollback

seed:
	go run ./cmd/game-library-auth-manage/. seed

keygen:
	go run ./cmd/game-library-auth-manage/. keygen

dockerbuildmng:
	docker build -f Dockerfile.mng -t game-library-auth-mng:latest .

dockerrunmng-m:
	docker compose run --rm mng migrate

dockerrunmng-r:
	docker compose run --rm mng rollback

dockerrunmng-s:
	docker compose run --rm mng seed
