# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -trimpath -o /bin/server cmd/server/main.go && \
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -trimpath -o /bin/seed   cmd/seed/main.go

# Runtime stage
FROM scratch

COPY --from=builder /bin/server /bin/server
COPY --from=builder /bin/seed   /bin/seed

ENV PORT=8080

EXPOSE 8080

USER 65534:65534

ENTRYPOINT ["/bin/server"]
