FROM golang:1.25 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o passkey-service .

FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=builder /app/passkey-service .
EXPOSE 8085
CMD ["./passkey-service"]
