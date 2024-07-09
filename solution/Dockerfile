FROM golang:1.22.1-alpine3.19 AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download && go mod verify

COPY . .

RUN go build -o ./bin/social ./cmd/main.go

FROM alpine:3.19 AS runner

WORKDIR /lib/social

COPY --from=builder /app/bin ./

RUN adduser -DH socialusr && chown -R socialusr: /lib/social && chmod -R 700 /lib/social

USER socialusr
 
CMD [ "./social" ]