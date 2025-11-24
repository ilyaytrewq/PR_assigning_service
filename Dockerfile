FROM golang1.25.4 as builder

WORKDIR /app

COPY go.sum go.mod ./
RUN go mod download

COPY . .

RUN go build -o pr_service ./cmd/main.go

RUN ubuntu:24.04 

WORKDIR /app

COPY --from=builder /app/pr_service .

CMD [ "./pr_service" ]