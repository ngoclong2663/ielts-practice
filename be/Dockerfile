# ---- build ----
FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/api ./cmd

# ---- migrate CLI ----
FROM migrate/migrate:v4.18.3 AS migrate

# ---- runtime ----
FROM alpine:3.22
RUN apk add --no-cache ca-certificates && adduser -D -H app
WORKDIR /app
COPY --from=build /out/api /usr/local/bin/api
COPY --from=migrate /usr/local/bin/migrate /usr/local/bin/migrate
COPY migrations ./migrations
COPY scripts/docker-entrypoint.sh /usr/local/bin/docker-entrypoint.sh
RUN chmod +x /usr/local/bin/docker-entrypoint.sh
USER app
ENV APP_ENV=production PORT=8080
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/docker-entrypoint.sh"]
