FROM golang:1.24-alpine3.22 AS build

LABEL authors="taking"

WORKDIR /app

COPY . .

RUN go build -o api-server ./cmd/api


FROM alpine:3.22 AS runtime-base

WORKDIR /app

COPY --from=build /app/api-server .
COPY --from=build /app/.env .

CMD ["./api-server"]