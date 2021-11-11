FROM golang:1.17.3-stretch AS builder
WORKDIR /gmcollector

COPY app/ ./app
COPY main.go ./
COPY go.mod ./ 
COPY go.sum ./

RUN ls && pwd 
RUN go mod download

RUN go build -o ./gmc


FROM debian:stable-slim
WORKDIR /app
COPY --from=builder /gmcollector/gmc ./gmc
COPY app/config.yaml ./config.yaml