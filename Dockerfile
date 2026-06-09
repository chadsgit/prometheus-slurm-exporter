FROM golang:1.25-alpine AS builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o prometheus-slurm-exporter ./cmd/exporter

FROM alpine:3.21
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /build/prometheus-slurm-exporter .
EXPOSE 8080
ENTRYPOINT ["/app/prometheus-slurm-exporter"]
