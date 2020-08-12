FROM alpine:3

# hadolint ignore=DL3018
RUN apk --update --no-cache add \
        git

COPY eclint /usr/local/bin/

CMD ["eclint"]
