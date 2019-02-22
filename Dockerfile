FROM alpine:3.7
MAINTAINER uSwitch Cloud <cloud@uswitch.com>

RUN  apk update && apk add ca-certificates && rm -rf /var/cache/apk/*

COPY bin/aws-signing-proxy-linux-amd64 /aws-signing-proxy

EXPOSE 8080

ENTRYPOINT ["/aws-signing-proxy"]
