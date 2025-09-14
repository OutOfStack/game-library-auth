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

SWAG_PKG := github.com/swaggo/swag/cmd/swag@v1.16.4
SWAG_BIN := $(shell go env GOPATH)/bin/swag
MOCKGEN_PKG := go.uber.org/mock/mockgen@v0.6
MOCKGEN_BIN := $(shell go env GOPATH)/bin/mockgen
generate:
	@if \[ ! -f ${SWAG_BIN} \]; then \
		echo "Installing swag..."; \
    	go install ${SWAG_PKG}; \
  	fi
	@if \[ -f ${SWAG_BIN} \]; then \
  		echo "Found swag at '$(SWAG_BIN)', generating documentation..."; \
	else \
    	echo "swag not found or the file does not exist"; \
    	exit 1; \
  	fi
	${SWAG_BIN} init \
	-d cmd/game-library-auth,internal/handlers,internal/web

	@if \[ ! -f ${MOCKGEN_BIN} \]; then \
		echo "Installing mockgen..."; \
		go install ${MOCKGEN_PKG}; \
	fi
	@if \[ -f ${MOCKGEN_BIN} \]; then \
		echo "Found mockgen at '$(MOCKGEN_BIN)', generating mocks..."; \
	else \
		echo "mockgen not found or the file does not exist"; \
		exit 1; \
  	fi
	${MOCKGEN_BIN} -source=internal/handlers/auth.go -destination=internal/handlers/mocks/auth.go -package=handlers_mocks
	${MOCKGEN_BIN} -source=internal/facade/provider.go -destination=internal/facade/mocks/provider.go -package=facade_mocks

LINT_PKG := github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.4
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
	go run ./cmd/game-library-auth-manage/. -from-file migrate

# rollback last migration
rollback:
	go run ./cmd/game-library-auth-manage/. -from-file rollback

keygen:
	go run ./cmd/game-library-auth-manage/. keygen

dbuildmng:
	docker build -f Dockerfile.mng -t game-library-auth-mng:latest .
