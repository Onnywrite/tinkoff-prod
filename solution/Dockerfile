FROM golang:1.22.0-alpine3.19 AS builder

RUN apk update && apk add ca-certificates git gcc g++ libc-dev binutils

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

RUN go build -o bin/application cmd/main.go

FROM alpine:3.19 AS runner

RUN apk update && apk add ca-certificates libc6-compat openssh bash && rm -rf /var/cache/apk/*

WORKDIR /app

COPY --from=builder /app/bin/application ./

COPY ./resources ./resources
ARG pg_conn
ARG server_addr=:8080
ENV POSTGRES_CONN=$pg_conn SERVER_ADDRESS=$server_addr

EXPOSE 8080:8080/tcp
CMD ["application"]