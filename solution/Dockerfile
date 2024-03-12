FROM golang:1.21-alpine3.19 AS builder

RUN apk update && apk add ca-certificates git gcc g++ libc-dev binutils

WORKDIR /opt

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

RUN go build -o bin/application cmd/main.go

FROM alpine:3.19 AS runner

RUN apk update && apk add ca-certificates libc6-compat openssh bash && rm -rf /var/cache/apk/*

WORKDIR /opt

COPY --from=builder /opt/bin/application ./

COPY ./resources ./resources
# --build-arg pg_conn=$POSTGRES_CONNECTION --build-arg server_addr=$SERVER_ADDRESS
# DOES NOT WORK
ARG pg_conn
ARG server_addr=8080
ENV POSTGRES_CONN=$pg_conn SERVER_ADDRESS=$server_addr

EXPOSE $server_addr
CMD ["./application"]