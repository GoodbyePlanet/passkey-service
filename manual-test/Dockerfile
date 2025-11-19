FROM golang:1.25 AS builder
WORKDIR /app
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o manual-test-server .

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/manual-test-server .
COPY --from=builder /app/static ./static
EXPOSE 6334
CMD ["./manual-test-server"]
