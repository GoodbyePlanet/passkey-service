FROM golang:1.25 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o passkey-service main.go

FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=builder /app/passkey-service .
EXPOSE 8080
CMD ["./passkey-service"]
