# Build context is the module root, apps/backend (issue #17).
FROM golang:1.27-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -o /bff ./cmd/bff

FROM alpine:3
COPY --from=build /bff /bff
EXPOSE 8080
ENTRYPOINT ["/bff"]
