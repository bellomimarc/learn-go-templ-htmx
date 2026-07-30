# syntax=docker/dockerfile:1

FROM golang:1.25-alpine AS builder

WORKDIR /src

# Cache module downloads first.
COPY go.mod go.sum ./
RUN go mod download

# Copy the full project and build a static Linux binary.
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags='-s -w' -o /out/saas-poc ./cmd/server

FROM scratch AS runtime

WORKDIR /

# Optional but useful if the app later performs outbound HTTPS calls.
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy only the compiled application.
COPY --from=builder /out/saas-poc /saas-poc

EXPOSE 8080

# Use an unprivileged UID/GID.
USER 10001:10001

ENTRYPOINT ["/saas-poc"]
