# Build context is the module root, apps/backend (issue #15).
FROM golang:1.27-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -o /members ./cmd/members

FROM alpine:3
COPY --from=build /members /members
EXPOSE 8080
ENTRYPOINT ["/members"]
