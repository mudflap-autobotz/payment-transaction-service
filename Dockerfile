# Stage 1 : build code
FROM golang:1.26-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

ARG GITHUB_USERNAME
ARG GITHUB_TOKEN

RUN git config --global url."https://${GITHUB_USERNAME}:${GITHUB_TOKEN}@github.com/".insteadOf "https://github.com/"

ENV GOPRIVATE="github.com/aaa-research/*"

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go tool swag init -g cmd/api/main.go -o docs --parseDependency --parseInternal --parseDepth 2

RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/api ./cmd/api

# Stage 2 : run appication
FROM alpine:latest

RUN apk --no-cache add ca-certificates

COPY --from=builder /bin/api /bin/api

EXPOSE 8080

ENTRYPOINT ["/bin/api"]
