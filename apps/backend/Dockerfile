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
COPY --from=build /app /app
EXPOSE 8080
ENTRYPOINT ["/app"]
