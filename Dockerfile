# To compile this image manually run:
#
# $ make docker
FROM gcr.io/distroless/static-debian11:debug-nonroot
COPY oathkeeper /usr/bin/oathkeeper
ENTRYPOINT ["oathkeeper"]
CMD ["serve"]
