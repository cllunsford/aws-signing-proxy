FROM alpine:3.7
MAINTAINER uSwitch Cloud <cloud@uswitch.com>

COPY bin/aws-signing-proxy-linux-amd64 /aws-signing-proxy
COPY entrypoint.sh /entrypoint.sh

EXPOSE 8080

ENTRYPOINT ["/entrypoint.sh"]
