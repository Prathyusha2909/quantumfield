FROM golang:1.23-alpine AS builder

ARG SERVICE=api
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/quantumfield ./cmd/${SERVICE}

FROM alpine:3.24

RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S quantumfield \
    && adduser -S -G quantumfield quantumfield

COPY --from=builder /out/quantumfield /usr/local/bin/quantumfield

USER quantumfield
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/quantumfield"]

