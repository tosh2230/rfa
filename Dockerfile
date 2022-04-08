FROM golang:1.16-buster as builder

ARG GCP_PROJECT_ID=""
WORKDIR /app
COPY go.* ./
RUN go mod download
COPY . ./
RUN go build ./cmd/server

FROM debian:buster-slim

RUN set -x && apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y \
    ca-certificates && \
    rm -rf /var/lib/apt/lists/*

# Copy the binary to the production image from the builder stage.
COPY --from=builder /app/server /app/server

# Run the web service on container startup.
CMD ["/app/server"]
