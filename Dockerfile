FROM golang:1.25.4-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o pr_service ./cmd/main.go

FROM alpine:3.20

WORKDIR /app

COPY --from=builder /app/pr_service .

EXPOSE 8080

CMD ["./pr_service"]
