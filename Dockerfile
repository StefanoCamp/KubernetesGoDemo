# syntax=docker/dockerfile:1

FROM golang:1.22-alpine AS build

WORKDIR /src

COPY go.mod ./
COPY main.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /app/kubernetes-go-api-demo .

FROM alpine:3.20

RUN addgroup -S app && adduser -S app -G app

WORKDIR /app

COPY --from=build /app/kubernetes-go-api-demo ./kubernetes-go-api-demo

USER app

EXPOSE 8080

ENV PORT=8080

ENTRYPOINT ["./kubernetes-go-api-demo"]
