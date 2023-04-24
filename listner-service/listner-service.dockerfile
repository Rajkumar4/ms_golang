FROM golang:1.18-alpine as builder

RUN mkdir /app

COPY . /app

WORKDIR /app

RUN CGO_ENABLED=0 go build -o listenApp .

RUN chmod +x /app/listenApp

# build a tiny docker image
FROM alpine:latest

RUN mkdir /app

COPY --from=builder /app/listenApp /app

CMD [ "/app/listenApp"]