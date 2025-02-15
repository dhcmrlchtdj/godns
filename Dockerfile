FROM gcr.io/distroless/static-debian12

LABEL org.opencontainers.image.source=https://github.com/dhcmrlchtdj/godns

COPY ./_build/godns /godns
ENTRYPOINT ["/godns"]
