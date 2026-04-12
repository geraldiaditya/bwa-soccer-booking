FROM golang:1.23.4-alpine3.21 as builder

RUN apk update
RUN apk add git openssh tzdata build-base python3 net-tools

WORKDIR /app

COPY .env.example .env
COPY . .

RUN go install github.com/buu700/gin@latest
RUN go mod tidy

RUN make build

FROM alpine:latest

RUN apk update && apk upgrade && \
    apk --update --no-cache add tzdata && \
    apk --no-cache add curl && \
    mkdir /app

WORKDIR /app

EXPOSE 8002

COPY --from=builder /app /app

ENTRYPOINT [ "/app/field-service" ]