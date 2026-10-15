FROM golang:1.26-alpine AS builder
WORKDIR /src
RUN apk add --no-cache ca-certificates git
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -tags nomsgpack -trimpath -ldflags='-s -w' -o /out/server ./services

FROM alpine:3.23 AS runner
RUN apk add --no-cache ca-certificates tzdata && addgroup -S -g 1001 blog && adduser -S -D -H -u 1001 -G blog blog
WORKDIR /app
COPY --from=builder --chown=blog:blog /out/server ./server
RUN mkdir -p /app/config /app/data && chown blog:blog /app/config /app/data
USER blog
EXPOSE 4000
ENTRYPOINT ["/app/server"]
CMD ["serve"]
