# build
FROM golang:1.26-alpine3.23 AS builder

WORKDIR /tmp/game-library-auth

# copy and download dependencies
COPY go.mod go.sum  ./
RUN go mod download

# copy code and config into container
COPY ./app.example.env ./out/app.env
COPY . .

# build app
ARG APP_VERSION=dev
ARG APP_COMMIT=unknown
RUN go build \
    -ldflags "-X github.com/OutOfStack/game-library-auth/internal/version.appVersion=${APP_VERSION} -X github.com/OutOfStack/game-library-auth/internal/version.appCommit=${APP_COMMIT}" \
    -o ./out/game-library-auth cmd/game-library-auth/main.go

# run
FROM alpine:3.23

WORKDIR /app

# copy built app into runnable container
COPY --from=builder /tmp/game-library-auth/out ./

EXPOSE 8000

ENTRYPOINT ["./game-library-auth"]
