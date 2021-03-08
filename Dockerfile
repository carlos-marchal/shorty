FROM golang:1.16.0-alpine3.13 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . ./
RUN CGO_ENABLED=0 GOOS=linux go build

FROM alpine:3.13.2
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /app/shorty .
CMD ["./shorty"]