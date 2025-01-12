# Builder
FROM golang:1.23-alpine3.21 as builder

WORKDIR /usr/app

RUN apk update && apk upgrade && \
    apk --update add git make

COPY go.mod go.sum ./
RUN go mod download

COPY . ./

RUN make build

# Distribution
FROM alpine:3.21

WORKDIR /usr/app

COPY --from=builder /usr/app/bin/rate-limiter /usr/app/bin/

ENV PATH="/usr/app/bin:${PATH}"

ENTRYPOINT [ "rate-limiter" ]
CMD [ "--help" ]
