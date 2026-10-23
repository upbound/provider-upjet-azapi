FROM alpine:3.24.3

ARG TARGETOS
ARG TARGETARCH

USER ${USER_ID}
EXPOSE 8080

ADD "bin/${TARGETOS}_${TARGETARCH}/provider" /usr/local/bin/provider

ENTRYPOINT ["provider"]
