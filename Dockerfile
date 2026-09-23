# Build stage
FROM golang:1.24-alpine AS builder

ENV CGO_ENABLED=0 \
    GOOS=linux

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -v -trimpath -ldflags="-s -w" -buildvcs=false -o expertlisting-server ./cmd/main.go

# Run stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata
RUN adduser -D appuser
WORKDIR /app

COPY --from=builder /app/expertlisting-server .

USER appuser
EXPOSE 8082

CMD ["./expertlisting-server"]
