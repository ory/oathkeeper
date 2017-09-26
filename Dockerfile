FROM golang:1.7-alpine

RUN apk add --no-cache git mercurial
RUN go get github.com/Masterminds/glide
WORKDIR /go/src/github.com/ory-am/editor-platform/services/exposed/proxies/firewall-reverse-proxy

ADD ./glide.lock ./glide.lock
ADD ./glide.yaml ./glide.yaml
RUN glide install

ADD . .
RUN go install .

ENTRYPOINT /go/bin/firewall-reverse-proxy

EXPOSE 3000