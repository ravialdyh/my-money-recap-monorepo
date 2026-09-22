FROM golang:1.27-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG CMD
RUN CGO_ENABLED=0 go build -trimpath -o /app ./cmd/$CMD

FROM alpine:3
RUN adduser -D -H -u 10001 app
COPY --from=build /app /app
USER 10001
EXPOSE 8080
ENTRYPOINT ["/app"]
