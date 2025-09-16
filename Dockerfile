FROM golang:1.25.1 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o web ./cmd/web

FROM gcr.io/distroless/base-debian12 AS runner

WORKDIR /app

COPY --from=builder /app/ /app/

EXPOSE 8000

CMD ["./web"]
