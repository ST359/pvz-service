FROM golang:1.24 AS builder

WORKDIR /app
COPY . ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /pvz-service ./cmd/pvz-service/ \
    && go clean -cache -modcache

FROM alpine:latest
WORKDIR /
COPY --from=builder /pvz-service ./pvz-service
RUN ls -l

EXPOSE 8080

CMD ["/pvz-service"]