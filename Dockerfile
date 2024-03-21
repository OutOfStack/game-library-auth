# build
FROM golang:1.22-alpine3.19 as builder

WORKDIR /tmp/game-library-auth

# copy and download dependencies
COPY go.mod go.sum  ./
RUN go mod download

# copy code and config into container
COPY ./app.env ./out/
COPY . .

# build app
RUN CGO_ENABLED=0 go build -o ./out/game-library-auth cmd/game-library-auth/main.go

# run
FROM alpine:3.19 AS build-release-stage

WORKDIR /app

# copy built app into runnable container
COPY --from=builder /tmp/game-library-auth/out ./

EXPOSE 8000

ENTRYPOINT ["./game-library-auth"]
