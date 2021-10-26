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