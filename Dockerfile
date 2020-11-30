FROM golang:1.15-alpine AS build_base

RUN apk add --no-cache git

# Set the Current Working Directory inside the container
WORKDIR /tmp/modsecurity-sentry

# We want to populate the module cache based on the go.{mod,sum} files.
COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

# Unit tests
# RUN CGO_ENABLED=0 go test -v

# Build the Go app
RUN go build -o ./out/modsecurity-sentry .

# Start fresh from a smaller image
FROM alpine:3.12 
RUN apk add ca-certificates

COPY --from=build_base /tmp/modsecurity-sentry/out/modsecurity-sentry /bin/modsecurity-sentry

# Run the binary program produced by `go install`
CMD ["modsecurity-sentry"]