FROM golang:1.26-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o main ./cmd/api

FROM alpine:latest
WORKDIR /app

COPY --from=build /app/main .
COPY config/config.example.yaml ./config.yaml

EXPOSE 8080

ENTRYPOINT ["./main", "-config", "/app/config.yaml"]
