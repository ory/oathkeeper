FROM alpine:3.6

RUN apk add --update ca-certificates # Certificates for SSL

ADD oathkeeper-docker-bin /go/bin/oathkeeper
RUN chmod a=+x /go/bin/oathkeeper

ENTRYPOINT ["/go/bin/oathkeeper"]
