# Stage 1 : build code
FROM golang:1.26-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

ENV GOPRIVATE="github.com/mudflap-autobotz/*"

COPY go.mod go.sum ./

RUN --mount=type=secret,id=github_token \
    export GIT_CONFIG_COUNT=1 && \
    export GIT_CONFIG_KEY_0="url.https://x-access-token:$(cat /run/secrets/github_token)@github.com/.insteadOf" && \
    export GIT_CONFIG_VALUE_0="https://github.com/" && \
    go mod download

COPY . .

RUN go tool swag init -g cmd/api/main.go -o docs --parseDependency --parseInternal --parseDepth 2

RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/api ./cmd/api

# Stage 2 : run appication
FROM alpine:latest

RUN apk --no-cache add ca-certificates

COPY --from=builder /bin/api /bin/api

EXPOSE 8080

ENTRYPOINT ["/bin/api"]
