# syntax=docker/dockerfile:1

# -------- STAGE 1: BUILD --------
FROM golang:1.21 AS builder

WORKDIR /app

# Use Docker cache effectively
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . ./

RUN CGO_ENABLED=0 GOOS=linux go build -o url-shortener .

# -------- STAGE 2: RUNTIME --------
FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/url-shortener .

EXPOSE 8080

CMD ["./url-shortener"]