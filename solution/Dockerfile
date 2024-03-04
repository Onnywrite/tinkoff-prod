FROM golang:1.22.0-alpine3.19 AS builder

RUN apk update && apk add ca-certificates git gcc g++ libc-dev binutils

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

RUN go build -o bin/application.elf cmd/main.go

FROM alpine:3.19 AS runner

RUN apk update && apk add ca-certificates libc6-compat openssh bash && rm -rf /var/cache/apk/*

WORKDIR /app

COPY --from=builder /app/bin/application.elf ./

COPY ./resources ./resources

EXPOSE 8080:8080/tcp
CMD ["./application.elf"]