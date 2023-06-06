FROM alpine:3.17.2
ARG TARGETARCH
COPY bin/collie-api-server-$TARGETARCH /usr/local/bin/collie-api-server
COPY assets /assets
COPY config/*.yaml /config/
CMD ["collie-api-server"]
