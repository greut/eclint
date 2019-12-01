FROM alpine:3

RUN apk --update add git libmagic && \
    rm -rf /var/lib/apt/lists/* && \
    rm /var/cache/apk/*

COPY eclint /usr/local/bin/

CMD ["eclint"]
