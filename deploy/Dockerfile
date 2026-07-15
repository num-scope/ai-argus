FROM golang:1.25-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/ai-argus ./cmd/server

FROM alpine:3.22
RUN addgroup -S argus && adduser -S -G argus argus
WORKDIR /app
COPY --from=builder /out/ai-argus /usr/local/bin/ai-argus
RUN mkdir -p /app/data && chown -R argus:argus /app
USER argus
ENV ARGUS_ADDRESS=:8080
ENV ARGUS_DATABASE_PATH=/app/data/ai-argus.db
EXPOSE 8080
VOLUME ["/app/data"]
ENTRYPOINT ["ai-argus"]
