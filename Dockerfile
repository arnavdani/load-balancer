FROM golang:1.22.0 AS builder
WORKDIR /app
COPY load-balancer.go go.mod ./
RUN CGO_ENABLED=0 GOOS=linux go build -o load-balancer .

FROM alpine:latest  
RUN apk --no-cache add ca-certificates
WORKDIR /root
COPY --from=builder /app/load-balancer .
ENTRYPOINT [ "/root/load-balancer" ]