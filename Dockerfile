# To compile this image manually run:
#
# $ GO111MODULE=on GOOS=linux GOARCH=amd64 go build && docker build -t oryd/oathkeeper . && rm oathkeeper
FROM alpine:3.10

RUN apk add -U --no-cache ca-certificates

COPY oathkeeper /usr/bin/oathkeeper

USER 1000

ENTRYPOINT ["oathkeeper"]
CMD ["serve"]
