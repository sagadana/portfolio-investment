
# Stage 1: Build the Go application
FROM golang:1.24.5-alpine AS builder
WORKDIR /usr/src/app

COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN apk add --no-cache make build-base
RUN go env -w CGO_ENABLED=1
RUN go build -v -o /usr/local/bin/app ./

# Stage 2: Create a minimal image with the built application
FROM alpine:latest AS runtime
RUN apk --no-cache add ca-certificates
COPY --from=builder /usr/local/bin/app /usr/local/bin/app
ENTRYPOINT ["app"]