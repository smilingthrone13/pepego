FROM golang:1.22-alpine AS build-stage

RUN apk update
RUN apk upgrade
RUN apk add build-base

WORKDIR /app

COPY . .

RUN go mod download

WORKDIR /app/cmd

RUN CGO_ENABLED=1 GOOS=linux go build -o pepobot main.go

FROM alpine:latest AS release-stage

WORKDIR /app

COPY --from=build-stage /app/cmd/pepobot .
COPY --from=build-stage /app/migrations ./migrations

RUN chmod +x pepobot

CMD ["./pepobot"]