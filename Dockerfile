FROM golang:1.16-alpine AS builder
ENV GO111MODULE=on \
    CGO_ENABLED=0
WORKDIR /src
ADD . .
RUN go build -o ./bin/wrss

FROM alpine:3.12
RUN apk add --no-cache tini
COPY --from=builder /src/bin/wrss /app/wrss
WORKDIR /app
ENTRYPOINT ["/sbin/tini", "--"]
CMD ["./wrss"]
