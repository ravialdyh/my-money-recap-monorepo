# One image recipe for every binary in this module; CMD picks the cmd/ folder (ARCHITECTURE §10).
# Build with the module root as context: docker build --build-arg CMD=bff .
FROM golang:1.27-alpine AS build
ARG CMD
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -o /app ./cmd/$CMD

FROM alpine:3
# Not root. The binary needs no privilege: it listens above 1024, writes nothing and reaches
# Postgres and Keycloak over the network only.
RUN adduser -D -H -u 10001 app
COPY --from=build /app /app
USER 10001
EXPOSE 8080
ENTRYPOINT ["/app"]
