FROM alpine:3.23.4

ARG TARGETOS
ARG TARGETARCH

USER ${USER_ID}
EXPOSE 8080

ADD "bin/${TARGETOS}_${TARGETARCH}/provider" /usr/local/bin/provider

ENTRYPOINT ["provider"]
