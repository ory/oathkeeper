# To compile this image manually run:
#
# $ make docker
FROM alpine:3.18 as base

RUN apk --no-cache --update-cache --upgrade --latest add ca-certificates

#############
FROM scratch

COPY --from=base /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY oathkeeper /usr/bin/oathkeeper

USER 1000

EXPOSE 4455
EXPOSE 4456

ENTRYPOINT ["oathkeeper"]
CMD ["serve"]
