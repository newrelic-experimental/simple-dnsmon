FROM golang:alpine
RUN apk update && apk add --no-cache git

RUN pwd
COPY . /DnsMon
WORKDIR /DnsMon

RUN go get -d -v

RUN go build -o /DnsMon/dnsmon
RUN ls /DnsMon/dnsmon

ENTRYPOINT ["/DnsMon/dnsmon"]

