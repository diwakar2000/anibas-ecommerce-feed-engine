FROM golang:1.24-alpine AS build

WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/api ./cmd/api

FROM alpine:3.21

RUN addgroup -S app && adduser -S app -G app && apk add --no-cache ca-certificates

USER app
WORKDIR /app

COPY --from=build /out/api /app/api

EXPOSE 8080
ENTRYPOINT ["/app/api"]

