# To compile this image manually run:
#
# $ make docker
FROM alpine:3.13

RUN apk add -U --no-cache ca-certificates

FROM scratch

COPY --from=0 /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY oathkeeper /usr/bin/oathkeeper

USER 1000

ENTRYPOINT ["oathkeeper"]
CMD ["serve"]
