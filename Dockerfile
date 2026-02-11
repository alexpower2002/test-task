FROM golang:latest AS build

ENV GO111MODULE=on

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /app/task-tracker ./cmd/task-tracker

FROM alpine:3.13

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=build /app/task-tracker /app/task-tracker
COPY --from=build /app/.env .env
COPY --from=build /app/internal/repository/mysql_migrations ./mysql_migrations

EXPOSE 8080

ENTRYPOINT ["/app/task-tracker"]
