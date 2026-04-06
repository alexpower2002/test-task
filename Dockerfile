FROM golang:1.25-alpine AS build

WORKDIR /src

RUN apk add --no-cache git

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /out/app ./cmd/app

FROM alpine:3.21

WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY --from=build /out/app ./app
COPY migrations ./migrations

EXPOSE 9001 9090

CMD ["./app"]
