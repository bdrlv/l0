FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY cmd/ cmd/
COPY internal/ internal/
# COPY .env .

# COPY migrations/init.sql migrations/init.sql 

RUN go build -o l0 ./cmd/main.go
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/l0 .
# COPY --from=builder /app/init.sql .

EXPOSE 8081

CMD ["./l0"]