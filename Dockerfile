ARG BASE=golang:1.16-alpine3.12
FROM ${BASE} AS builder

RUN sed -e 's/dl-cdn[.]alpinelinux.org/nl.alpinelinux.org/g' -i~ /etc/apk/repositories
RUN apk add --update --no-cache make git openssh gcc libc-dev zeromq-dev libsodium-dev

# set the working directory
WORKDIR /device-tuya-go

COPY . .

RUN go mod tidy
RUN go mod download

RUN make build

FROM alpine:3.12

RUN sed -e 's/dl-cdn[.]alpinelinux.org/nl.alpinelinux.org/g' -i~ /etc/apk/repositories
RUN apk add --update --no-cache zeromq dumb-init

COPY --from=builder /device-tuya-go/cmd /

EXPOSE 59988

ENTRYPOINT ["/device-tuya"]
CMD ["--confdir=/res"]
