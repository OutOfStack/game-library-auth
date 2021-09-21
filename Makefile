build:
	mkdir -p bin
	go build -o bin/game-library-auth cmd/game-library-auth/main.go

run:
	go run ./cmd/game-library-auth/.

dockerrunpg:
	docker-compose up -d db

createdb:
	docker exec -it auth_db createdb --username=postgres --owner=postgres auth

dropdb:
	docker exec -it auth_db dropdb --username=postgres auth