### Build stage
FROM golang:1.25 AS builder

WORKDIR /build

# Install just and swag for build orchestration and swagger doc generation.
RUN go install github.com/swaggo/swag/cmd/swag@latest && \
    curl --proto '=https' --tlsv1.2 -sSf https://just.systems/install.sh | bash -s -- --to /usr/local/bin

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary and generate swagger docs using just so build commands
# are maintained in one place (the Justfile).
ENV CGO_ENABLED=0

RUN just ldflags='-w -s' build

### Final stage - minimal runtime image
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /

# Copy the binary from builder
COPY --from=builder /build/bin/analyses /analyses

# Copy swagger documentation
COPY --from=builder /build/docs/swagger.json /docs/swagger.json

EXPOSE 60000

ENTRYPOINT ["/analyses"]
