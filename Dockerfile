FROM golang:alpine as builder

LABEL stage=gobuilder
# RUN apk update --no-cache && apk add --no-cache tzdata

WORKDIR /build

ADD go.mod .
ADD go.sum .
RUN go mod download

COPY . .
RUN go build -o /app/epstgbot .


FROM alpine:latest
USER root

COPY --from=builder /app/epstgbot /usr/local/bin/epstgbot

ENTRYPOINT ["epstgbot"]
