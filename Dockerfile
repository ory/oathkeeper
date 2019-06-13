# To compile this image manually run:
#
# $ GO111MODULE=on GOOS=linux GOARCH=amd64 go build && docker build -t oryd/oathkeeper . && rm oathkeeper
FROM alpine:3.9

RUN apk add -U --no-cache ca-certificates

FROM scratch

COPY --from=0 /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY oathkeeper /usr/bin/oathkeeper

ENTRYPOINT ["oathkeeper"]
