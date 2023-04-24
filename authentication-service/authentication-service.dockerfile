FROM golang:1.18-alpine as builder

RUN mkdir /app

COPY . /app

WORKDIR /app

RUN CGO_ENABLED=0 go build -o authapp ./cmd/api

RUN chmod +x /app/authapp

# build a tiny docker image
FROM alpine:latest

RUN mkdir /app

COPY --from=builder /app/authapp /app

CMD [ "/app/authapp" ]