FROM golang:1.27-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -o /out/ ./cmd/...

FROM alpine:3
RUN adduser -D -H -u 10001 app
ARG CMD
COPY --from=build /out/${CMD:?} /app
USER 10001
EXPOSE 8080
ENTRYPOINT ["/app"]
